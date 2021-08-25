package main

import (
	"os"

	tmCli "github.com/arcology/3rd-party/tm/cli"
	"github.com/arcology/eshing-svc/service"
)

func main() {
	st := service.StartCmd

	cmd := tmCli.PrepareMainCmd(st, "BC", os.ExpandEnv("$HOME/monacos/eshing"))
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
