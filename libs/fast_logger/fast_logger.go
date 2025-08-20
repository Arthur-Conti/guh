package fl

import (
	applicationpackage "github.com/Arthur-Conti/guh/libs/log/application_package"
	"github.com/Arthur-Conti/guh/libs/log/logger"
	"github.com/Arthur-Conti/guh/libs/log/outputs"
)

func Log(message string) {
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
	l := logger.NewLogger(loggerOpts)
	l.Debug(logger.LogMessage{ApplicationPackage: "fast_logger", Message: message})
}

func Logf(message string, args ...any) {
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
	l := logger.NewLogger(loggerOpts)
	l.Debugf(logger.LogMessage{ApplicationPackage: "fast_logger", Message: message, Vals: args})
}