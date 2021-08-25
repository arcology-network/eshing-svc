package service

import (
	"net/http"

	"github.com/arcology/component-lib/actor"
	"github.com/arcology/component-lib/aggregator"
	mworkers "github.com/arcology/component-lib/intl/module/workers"
	"github.com/arcology/component-lib/storage"
	"github.com/arcology/component-lib/streamer"
	"github.com/arcology/eshing-svc/service/workers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"

	//"github.com/sirupsen/logrus"
	"github.com/arcology/component-lib/kafka"
)

type Config struct {
	concurrency int
	groupid     string
}

//return a Subscriber struct
func NewConfig() *Config {
	return &Config{
		concurrency: viper.GetInt("concurrency"),
		groupid:     "eshing",
	}
}

func (cfg *Config) Start() {
	// logrus.SetLevel(logrus.DebugLevel)
	// logrus.SetFormatter(&logrus.JSONFormatter{})

	http.Handle("/streamer", promhttp.Handler())
	go http.ListenAndServe(":19004", nil)

	broker := streamer.NewStatefulStreamer()
	//00 initializer
	initializer := actor.NewActor(
		"initializer",
		broker,
		[]string{actor.MsgStarting},
		[]string{
			actor.MsgStartSub,
		},
		[]int{1},
		workers.NewInitializer(cfg.concurrency, cfg.groupid),
	)
	initializer.Connect(streamer.NewDisjunctions(initializer, 1))

	receiveMseeages := []string{
		actor.MsgInitExecuted,
		actor.MsgInclusive,
		actor.MsgAppHash,
	}

	receiveTopics := []string{
		viper.GetString("msgexch"),
		viper.GetString("genesis-apc"),
		viper.GetString("inclusive-txs"),
	}

	//01 kafkaDownloader
	kafkaDownloader := actor.NewActor(
		"kafkaDownloader",
		broker,
		[]string{actor.MsgStartSub},
		receiveMseeages,
		[]int{1, 1, 1},
		kafka.NewKafkaDownloader(cfg.concurrency, cfg.groupid, receiveTopics, receiveMseeages, viper.GetString("mqaddr")),
	)
	kafkaDownloader.Connect(streamer.NewDisjunctions(kafkaDownloader, 5))

	//01 kafkaDownloader
	kafkaDownloaderEu := actor.NewActor(
		"kafkaDownloaderEu",
		broker,
		[]string{actor.MsgStartSub},
		[]string{actor.MsgEuResults},
		[]int{100},
		kafka.NewKafkaDownloader(cfg.concurrency, cfg.groupid, []string{viper.GetString("euresults")}, []string{actor.MsgEuResults}, viper.GetString("mqaddr2")),
	)
	kafkaDownloaderEu.Connect(streamer.NewDisjunctions(kafkaDownloaderEu, 100))

	//02 hpmtworker
	hpmtWorker := actor.NewActor(
		"hpmtWorker",
		broker,
		[]string{
			actor.MsgExecuted,
		},
		[]string{
			actor.MsgAcctHash,
			actor.MsgInitParentInfo,
			actor.MsgClearCompletedEuresults,
		},
		[]int{1, 1, 1},
		workers.NewHpmtWorker(cfg.concurrency, cfg.groupid),
	)
	hpmtWorker.Connect(streamer.NewDisjunctions(hpmtWorker, 1))

	//03 aggre
	euresultAggreSelector := actor.NewActor(
		"euresultAggreSelector",
		broker,
		[]string{
			actor.MsgEuResults,
			actor.MsgInclusiveEuresults,
			actor.MsgClearEuresults,
		},
		[]string{
			actor.MsgExecuted,
			actor.MsgSelectCompletedEuresults,
			actor.MsgClearCompletedEuresults,
		},
		[]int{1, 1, 1},
		mworkers.NewEuResultsAggreSelector(cfg.concurrency, cfg.groupid),
	)
	euresultAggreSelector.Connect(streamer.NewDisjunctions(euresultAggreSelector, 1))

	//04-1 directorClearEuresults
	relationClearEuresults := map[string]string{}
	relationClearEuresults[actor.MsgAppHash] = actor.MsgClearEuresults
	directorClearEuresults := actor.NewActor(
		"directorClearEuresults",
		broker,
		[]string{
			actor.MsgSelectCompletedEuresults,
			actor.MsgAppHash,
		},
		[]string{
			actor.MsgClearEuresults,
		},
		[]int{1},
		aggregator.NewDirector(cfg.concurrency, cfg.groupid, relationClearEuresults),
	)
	directorClearEuresults.Connect(streamer.NewConjunctions(directorClearEuresults))

	//04-2 directorListEuresults
	relationListEuresults := map[string]string{}
	relationListEuresults[actor.MsgInclusive] = actor.MsgInclusiveEuresults
	directorListEuresults := actor.NewActor(
		"directorListEuresults",
		broker,
		[]string{
			actor.MsgClearCompletedEuresults,
			actor.MsgInclusive,
		},
		[]string{
			actor.MsgInclusiveEuresults,
		},
		[]int{1},
		aggregator.NewDirector(cfg.concurrency, cfg.groupid, relationListEuresults),
	)
	directorListEuresults.Connect(streamer.NewConjunctions(directorListEuresults))

	relations := map[string]string{}
	relations[actor.MsgAcctHash] = viper.GetString("msgexch")
	relations[actor.MsgInitParentInfo] = viper.GetString("msgexch")

	//05 kafkaUploader
	kafkaUploader := actor.NewActor(
		"kafkaUploader",
		broker,
		[]string{
			actor.MsgAcctHash,
			actor.MsgInitParentInfo,
		},
		[]string{},
		[]int{},
		kafka.NewKafkaUploader(cfg.concurrency, cfg.groupid, relations, viper.GetString("mqaddr")),
	)
	kafkaUploader.Connect(streamer.NewDisjunctions(kafkaUploader, 4))

	//06 stateSwitcher
	stateSwitcher := actor.NewActor(
		"stateSwitcher",
		broker,
		[]string{
			actor.MsgInitExecuted,
		},
		[]string{actor.MsgExecuted},
		[]int{1},
		storage.NewSwitcher(cfg.concurrency, cfg.groupid, actor.MsgInitExecuted, actor.MsgExecuted),
	)
	stateSwitcher.Connect(streamer.NewDisjunctions(stateSwitcher, 1))

	//starter
	selfStarter := streamer.NewDefaultProducer("selfStarter", []string{actor.MsgStarting}, []int{1})
	broker.RegisterProducer(selfStarter)

	broker.Serve()

	//start signel
	streamerStarting := actor.Message{
		Name:   actor.MsgStarting,
		Height: 0,
		Round:  0,
		Data:   "start",
	}
	broker.Send(actor.MsgStarting, &streamerStarting)
}

func (cfg *Config) Stop() {

}
