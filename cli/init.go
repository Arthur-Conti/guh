package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Arthur-Conti/guh/config"
	"github.com/Arthur-Conti/guh/libs/db"
	errorhandler "github.com/Arthur-Conti/guh/libs/error_handler"
	"github.com/Arthur-Conti/guh/libs/log/logger"
	projectconfig "github.com/Arthur-Conti/guh/libs/project_config"
)

func Init() error {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	serviceName := fs.String("serviceName", "", "The name of the service")
	dbName := fs.String("dbName", "postgres", "The name of the database")
	github := fs.String("github", "", "The github url to init your go mod")
	gin := fs.Bool("gin", false, "Download the gin package")
	all := fs.Bool("all", false, "Create all files")
	showFirst := fs.Bool("showFirst", false, "Show a summary of everything that is gonna be created")
	help := fs.Bool("help", false, "Show help")
	fs.Parse(os.Args[2:])

	if *help {
		HelpInit()
	}

	if *serviceName == "" {
		return errorhandler.New(errorhandler.KindInvalidArgument, "Service name must be passed", errorhandler.WithOp("init"))
	}

	modName := *serviceName
	if *github != "" {
		modName = *github
	}

	if *showFirst {
		showStructure()
		fmt.Print("Continue? [Y/n]: ")
		var resp string
		fmt.Scanln(&resp)
		if resp != "" && strings.ToLower(resp) != "y" {
			config.Config.Logger.Warning(logger.LogMessage{ApplicationPackage: "cli", Message: "Aborted by user"})
			return nil
		}

		if err := createStructure(*serviceName, modName); err != nil {
			return errorhandler.Wrap(errorhandler.KindInternal, "failed to create structure", err, errorhandler.WithOp("init"))
		}
	} else {
		if err := createStructure(*serviceName, modName); err != nil {
			return errorhandler.Wrap(errorhandler.KindInternal, "failed to create structure", err, errorhandler.WithOp("init"))
		}
	}

	cfg, err := projectconfig.Load()
	if err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "failed to load project config", err, errorhandler.WithOp("init"))
	}
	cfg.ServiceName = *serviceName
	if err := projectconfig.Save(cfg); err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "failed to save project config", err, errorhandler.WithOp("init"))
	}

	if *all {
		if err := AllConfigs(); err != nil {
			return errorhandler.Wrap(errorhandler.KindInternal, "failed to create all configs", err, errorhandler.WithOp("init"))
		}
	}

	switch *dbName {
	case "postgres":
		if err := PostgresDockerCompose(*db.GetDefaultPostgresOpts()); err != nil {
			return errorhandler.Wrap(errorhandler.KindInternal, "failed to create postgres docker compose", err, errorhandler.WithOp("init"))
		}
	}

	if err := AddAppService("./docker-compose.yml", *serviceName); err != nil {
		return errorhandler.Wrap(errorhandler.KindInternal, "failed to add app service", err, errorhandler.WithOp("init"))
	}

	if *github != "" {
		if err := syncGithub(*github); err != nil {
			return errorhandler.Wrap(errorhandler.KindInternal, "failed to sync github", err, errorhandler.WithOp("init"))
		}
	} else {
		if err := modConfiguration(modName); err != nil {
			return errorhandler.Wrap(errorhandler.KindInternal, "failed to configure go mod", err, errorhandler.WithOp("init"))
		}
	}
	
	if *gin {
		if err := ginDownload(); err != nil {
			return errorhandler.Wrap(errorhandler.KindInternal, "failed to download gin", err, errorhandler.WithOp("init"))
		}
	}


	return nil
}

func HelpInit() {
	fmt.Println(`init - The init command help you initializing your project

Usage:
  guh init [flags]

Flags:
  --serviceName    The name of the service
  --dbName         The name of the database
  --github         The github url to init your go mod
  --gin            Download the gin package
  --all            Create all files
  --showFirst      Show a summary of everything that is gonna be created
  --help    Show help

Examples:
  guh init

For more information, visit: https://github.com/Arthur-Conti/guh`)
	os.Exit(0)
}