package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/Arthur-Conti/guh/config"
	errorhandler "github.com/Arthur-Conti/guh/libs/error_handler"
	"github.com/Arthur-Conti/guh/libs/log/logger"
)

var filePath *string

// Use a safe default path that can be overridden by flags
var configFilePath string = "./internal/config/"

func Config() error {
	fs := flag.NewFlagSet("config", flag.ExitOnError)
	// Local flag var; we will copy it to the global string after parsing
	filePath = fs.String("filePath", configFilePath, "The path where the file should go")
	all := fs.Bool("all", false, "Create all configs")
	logger := fs.Bool("logger", false, "Create logger config")
	help := fs.Bool("help", false, "Help with config command")
	fs.Parse(os.Args[2:])

	// Ensure global path is always set, even when other commands call these funcs directly
	if filePath != nil {
		configFilePath = *filePath
	}

	if *help {
		HelpConfig()
	}

	if *all {
		return AllConfigs()
	}
	if *logger {
		return LoggerConfig()
	}

	return nil
}

func AllConfigs() error {
	config.Config.Logger.Info(logger.LogMessage{ApplicationPackage: "cli", Message: "Creating all config files"})

	if err := InitConfig(); err != nil {
		return err
	}
	if err := LoggerConfig(); err != nil {
		return err
	}

	return nil
}

func InitConfig() error {
	config.Config.Logger.Info(logger.LogMessage{ApplicationPackage: "cli", Message: "Creating init config file"})
	fileName := "init.go"
	content := `package config

import "github.com/Arthur-Conti/guh/libs/log/logger"

type BaseConfigs struct {
	Logger *logger.Logger
}

var Config = &BaseConfigs{}

func Init() {
	Config.Logger = InitLogger()
}`

	return createFiles(configFilePath+fileName, content)
}

func LoggerConfig() error {
	config.Config.Logger.Info(logger.LogMessage{ApplicationPackage: "cli", Message: "Creating logger config file"})
	fileName := "logger.go"
	content := `package config

import (
	applicationpackage "github.com/Arthur-Conti/guh/libs/log/application_package"
	"github.com/Arthur-Conti/guh/libs/log/logger"
	"github.com/Arthur-Conti/guh/libs/log/outputs"
)

func InitLogger() *logger.Logger {
	plainOpts := outputs.PlainOutputOpts{
		DebugPattern:   "DEBUG: ",
		WarningPattern: "WARNING: ",
		InfoPattern:    "INFO: ",
		ErrorPattern:   "ERROR: ",
	}
	plainOutput := outputs.NewPlainOutput(plainOpts)
	appPackage := applicationpackage.NewPackageLevel()
	loggerOpts := logger.LoggerOpts{
		OutputType:         plainOutput,
		LevelStr:           "debug",
		ApplicationPackage: *appPackage,
	}
	return logger.NewLogger(loggerOpts)
}`

	return createFiles(configFilePath+fileName, content)
}

func createFiles(filePath, content string) error {
	if _, err := os.Stat(filePath); err == nil {
		config.Config.Logger.Infof(logger.LogMessage{ApplicationPackage: "cli", Message: "File '%s' already exists. Skipping creation.", Vals: []any{filePath}})
		return nil
	} else if !os.IsNotExist(err) {
		config.Config.Logger.Infof(logger.LogMessage{ApplicationPackage: "cli", Message: "Error checking file: %v\n", Vals: []any{err}})
		return errorhandler.Wrap(errorhandler.KindInternal, "Error checking file", err, errorhandler.WithOp("cli.createFiles"))
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error writing file: %v\n", Vals: []any{err}})
		return errorhandler.Wrap(errorhandler.KindInternal, "Error writing file", err, errorhandler.WithOp("cli.createFiles"))
	}
	config.Config.Logger.Infof(logger.LogMessage{ApplicationPackage: "cli", Message: "%v generated successfully.", Vals: []any{filePath}})

	return nil
}

func HelpConfig() {
	fmt.Println(`config - The config command help you creating your config files to you 

Usage:
  guh config [flags]

Flags:
  --filePath       Define where the config files must be created (Defaults to ./internal/config/)
  --all            If true creates all config files
  --logger         Creates only the logger file
  
Examples:
  guh compose --filePath=./
  guh compose --all
  guh compose --filePath=./config/ --logger

For more information, visit: https://github.com/Arthur-Conti/guh`)
	os.Exit(0)
}
