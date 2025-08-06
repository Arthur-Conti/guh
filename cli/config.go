package cli

import (
	"flag"
	"os"

	"github.com/Arthur-Conti/guh/config"
	"github.com/Arthur-Conti/guh/packages/log/logger"
)

var filePath string

func Config() error {
	fs := flag.NewFlagSet("compose", flag.ExitOnError)
	filePath = *fs.String("filePath", "./internal/main/", "The path where the file should go")
	all := fs.Bool("all", false, "Create all configs")
	logger := fs.Bool("logger", false, "Create logger config")
	fs.Parse(os.Args[2:])

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
	content := `
		package config

import "github.com/Arthur-Conti/logger/logger"

type BaseConfigs struct {
	Logger *logger.Logger
}

var Config = &BaseConfigs{}

func Init() {
	Config.Logger = InitLogger()
}`

	return createFiles(filePath+fileName, content)
}

func LoggerConfig() error {
	config.Config.Logger.Info(logger.LogMessage{ApplicationPackage: "cli", Message: "Creating logger config file"})
	fileName := "logger.go"
	content := `
	package config

import (
	applicationpackage "github.com/Arthur-Conti/logger/application_package"
	"github.com/Arthur-Conti/logger/logger"
	"github.com/Arthur-Conti/logger/outputs"
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

	return createFiles(filePath+fileName, content)
}

func createFiles(filePath, content string) error {
	if _, err := os.Stat(filePath); err == nil {
		config.Config.Logger.Infof(logger.LogMessage{ApplicationPackage: "cli", Message: "File '%s' already exists. Skipping creation.", Vals: []any{filePath}})
		return nil
	} else if !os.IsNotExist(err) {
		config.Config.Logger.Infof(logger.LogMessage{ApplicationPackage: "cli", Message: "Error checking file: %v\n", Vals: []any{err}})
		return err
	}

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "cli", Message: "Error writing file: %v\n", Vals: []any{err}})
		return err
	}

	config.Config.Logger.Infof(logger.LogMessage{ApplicationPackage: "cli", Message: "%v generated successfully.", Vals: []any{filePath}})

	return nil
}
