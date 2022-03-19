package service

import (
	"github.com/arcology-network/component-lib/actor"
	"github.com/arcology-network/component-lib/storage"
	"github.com/arcology-network/component-lib/streamer"
	"github.com/arcology-network/eshing-svc/service/workers"
)

type KafkaDownloaderCreator func(concurrency int, groupId string, topics, messageTypes []string, mqaddr string) actor.IWorker
type KafkaUploaderCreator func(concurrency int, groupId string, messages map[string]string, mqaddr string) actor.IWorker

type ConfigV2 struct {
	concurrency       int
	groupId           string
	topicMsgExch      string
	topicGenesisApc   string
	topicInclusiveTxs string
	topicEuResults    string
	kafka1            string
	kafka2            string
	kdc               KafkaDownloaderCreator
	kuc               KafkaUploaderCreator
	downloader1       actor.IWorker
	downloader2       actor.IWorker
	uploader          actor.IWorker
}

//return a Subscriber struct
func NewConfigV2(
	concurrency int,
	topicMsgExch string,
	topicGenesisApc string,
	topicInclusiveTxs string,
	topicEuResults string,
	kafka1 string,
	kafka2 string,
	kdc KafkaDownloaderCreator,
	kuc KafkaUploaderCreator,
) *ConfigV2 {
	return &ConfigV2{
		concurrency:       concurrency,
		groupId:           "eshing",
		topicMsgExch:      topicMsgExch,
		topicGenesisApc:   topicGenesisApc,
		topicInclusiveTxs: topicInclusiveTxs,
		topicEuResults:    topicEuResults,
		kafka1:            kafka1,
		kafka2:            kafka2,
		kdc:               kdc,
		kuc:               kuc,
	}
}

func (cfg *ConfigV2) Start() {
	broker := streamer.NewStatefulStreamer()

	initializer := actor.NewActor(
		"initializer",
		broker,
		[]string{actor.MsgStarting},
		[]string{
			actor.MsgStartSub,
			actor.MsgBlockCompleted,
		},
		[]int{1, 1},
		workers.NewInitializer(cfg.concurrency, cfg.groupId),
	)
	initializer.Connect(streamer.NewDisjunctions(initializer, 1))

	receiveMsgs := []string{
		actor.MsgInitExecuted,
		actor.MsgInclusive,
		actor.MsgAppHash,
		actor.MsgBlockCompleted,
	}
	receiveTopics := []string{
		cfg.topicMsgExch,
		cfg.topicGenesisApc,
		cfg.topicInclusiveTxs,
	}
	cfg.downloader1 = cfg.kdc(cfg.concurrency, cfg.groupId, receiveTopics, receiveMsgs, cfg.kafka1)
	kafkaDownloader := actor.NewActor(
		"kafkaDownloader",
		broker,
		[]string{actor.MsgStartSub},
		receiveMsgs,
		[]int{1, 1, 1, 1},
		cfg.downloader1,
	)
	kafkaDownloader.Connect(streamer.NewDisjunctions(kafkaDownloader, 5))

	cfg.downloader2 = cfg.kdc(cfg.concurrency, cfg.groupId, []string{cfg.topicEuResults}, []string{actor.MsgEuResults}, cfg.kafka2)
	kafkaDownloaderEu := actor.NewActor(
		"kafkaDownloaderEu",
		broker,
		[]string{actor.MsgStartSub},
		[]string{actor.MsgEuResults},
		[]int{100},
		cfg.downloader2,
	)
	kafkaDownloaderEu.Connect(streamer.NewDisjunctions(kafkaDownloaderEu, 100))

	rootCalculator := storage.NewDBHandler(cfg.concurrency, cfg.groupId, workers.NewRootCalculator())
	aggrSelector := workers.NewAggrSelector(cfg.concurrency, cfg.groupId)
	baseWorker := &actor.HeightController{}
	baseWorker.Next(&actor.FSMController{}).Next(actor.MakeLinkable(rootCalculator)).Next(&actor.FSMController{}).EndWith(aggrSelector)
	chainActor := actor.NewActor(
		"calculator-aggr-selector-chain",
		broker,
		[]string{
			actor.MsgEuResults,
			actor.MsgInclusive,
			actor.MsgExecuted,
			actor.MsgBlockCompleted,
		},
		[]string{
			actor.MsgExecuted,
			actor.MsgAcctHash,
			actor.MsgInitParentInfo,
		},
		[]int{1, 1, 1},
		baseWorker,
	)
	chainActor.Connect(streamer.NewDisjunctions(chainActor, 1))

	relations := map[string]string{
		actor.MsgAcctHash:       cfg.topicMsgExch,
		actor.MsgInitParentInfo: cfg.topicMsgExch,
	}
	cfg.uploader = cfg.kuc(cfg.concurrency, cfg.groupId, relations, cfg.kafka1)
	kafkaUploader := actor.NewActor(
		"kafkaUploader",
		broker,
		[]string{
			actor.MsgAcctHash,
			actor.MsgInitParentInfo,
		},
		[]string{},
		[]int{},
		cfg.uploader,
	)
	kafkaUploader.Connect(streamer.NewDisjunctions(kafkaUploader, 4))

	actor.Rename(actor.MsgInitExecuted).To(actor.MsgExecuted).On(broker)

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

func (cfg *ConfigV2) Stop() {

}
