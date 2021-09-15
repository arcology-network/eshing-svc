package main

import (
	"os"

	tmCli "github.com/arcology-network/3rd-party/tm/cli"
	"github.com/arcology-network/eshing-svc/service"
)

func main() {
	//debug.SetGCPercent(-1)
	st := service.StartCmd

	cmd := tmCli.PrepareMainCmd(st, "BC", os.ExpandEnv("$HOME/monacos/eshing"))
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
