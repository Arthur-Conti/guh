package main

import (
	"os"

	"github.com/Arthur-Conti/guh/cli"
	"github.com/Arthur-Conti/guh/config"
	"github.com/Arthur-Conti/guh/libs/log/logger"
)

func main() {
	config.Init()
	if len(os.Args) > 1 {
		if err := cli.Handle(os.Args); err != nil {
			panic(err)
		}
	} else {
		config.Config.Logger.Warning(logger.LogMessage{ApplicationPackage: "main", Message: "GUH DIDN'T FIND ANY COMMANDS"})
		cli.Help()
	}
}
