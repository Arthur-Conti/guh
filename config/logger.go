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
}