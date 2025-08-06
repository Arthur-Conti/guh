package config

import (
	applicationpackage "github.com/Arthur-Conti/guh/packages/log/application_package"
	"github.com/Arthur-Conti/guh/packages/log/logger"
	"github.com/Arthur-Conti/guh/packages/log/outputs"
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