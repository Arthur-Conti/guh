package main

import (
	"os"

	"github.com/Arthur-Conti/guh/cli"
	"github.com/Arthur-Conti/guh/config"
	"github.com/Arthur-Conti/guh/libs/log/logger"
)

func main() {
	config.Init()
	config.Config.Logger.Info(logger.LogMessage{ApplicationPackage: "main", Message: "Starting GUH"})
	if len(os.Args) > 1 {
		config.Config.Logger.Info(logger.LogMessage{ApplicationPackage: "main", Message: "GUH CLI MODE"})
		if err := cli.Handle(os.Args); err != nil {
			panic(err)
		}
	} else {
		config.Config.Logger.Info(logger.LogMessage{ApplicationPackage: "main", Message: "GUH DIDN'T FIND ANY COMMANDS"})
	}
}
