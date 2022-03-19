package service

import (
	"net/http"

	tmCommon "github.com/arcology-network/3rd-party/tm/common"
	"github.com/arcology-network/component-lib/kafka"
	"github.com/arcology-network/component-lib/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start eshing service Daemon",
	RunE:  startCmdV2,
}

func init() {
	flags := StartCmd.Flags()

	flags.String("mqaddr", "localhost:9092", "host:port of kafka ")
	flags.String("mqaddr2", "localhost:9092", "host:port of kafka ")

	//apc
	flags.String("genesis-apc", "genesis-apc", "topic for received genesis apc")

	//common
	flags.String("msgexch", "msgexch", "topic for receive msg exchange")
	flags.String("log", "log", "topic for send log")
	flags.Int("concurrency", 4, "num of threads")

	flags.String("inclusive-txs", "inclusive-txs", "topic of send txlist")
	flags.String("euresults", "euresults", "topic for received euresults")

	// flags.String("cfga", "/mnt/work/dataa/config.json", "cfg file of hpmt for  accounts ")
	// flags.String("cfgs", "/mnt/work/datas/config.json", "cfg file of hpmt for  account storages ")

	flags.String("logcfg", "./log.toml", "log conf path")

	// flags.Uint64("svcid", 8, "service id of eshing,range 1 to 255")
	// flags.Uint64("insid", 1, "instance id of eshing,range 1 to 255")

	// flags.Bool("draw", false, "draw flow graph")

	flags.Int("nidx", 0, "node index in cluster")
	flags.String("nname", "node1", "node name in cluster")
}

func startCmdV2(cmd *cobra.Command, args []string) error {
	log.InitLog("eshing.log", viper.GetString("logcfg"), "eshing", viper.GetString("nname"), viper.GetInt("nidx"))
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":19004", nil)

	en := NewConfigV2(
		viper.GetInt("concurrency"),
		viper.GetString("msgexch"),
		viper.GetString("genesis-apc"),
		viper.GetString("inclusive-txs"),
		viper.GetString("euresults"),
		viper.GetString("mqaddr"),
		viper.GetString("mqaddr2"),
		kafka.NewKafkaDownloader,
		kafka.NewKafkaUploader,
	)
	en.Start()

	// Wait forever
	tmCommon.TrapSignal(func() {
		// Cleanup
		en.Stop()
	})

	return nil
}
