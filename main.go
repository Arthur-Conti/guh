package main

import (
	"os"

	"github.com/Arthur-Conti/guh/cli"
	"github.com/Arthur-Conti/guh/config"
	"github.com/Arthur-Conti/guh/packages/db"
	"github.com/Arthur-Conti/logger/logger"
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
		config.Config.Logger.Info(logger.LogMessage{ApplicationPackage: "main", Message: "GUH PACKAGE MODE"})		
		_, err := db.DefaultPostgres()
		if err != nil {
			config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "main", Message: "Error connecting to postgres: %v", Vals: []any{err}})
		}
	}
}