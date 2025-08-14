package fl

import (
	"github.com/Arthur-Conti/guh/config"
	"github.com/Arthur-Conti/guh/libs/log/logger"
)

func Log(message string) {
	config.Config.Logger.Debug(logger.LogMessage{ApplicationPackage: "fast_logger", Message: message})
}

func Logf(message string, args ...any) {
	config.Config.Logger.Debugf(logger.LogMessage{ApplicationPackage: "fast_logger", Message: message, Vals: args})
}