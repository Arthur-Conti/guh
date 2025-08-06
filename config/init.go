package config

import "github.com/Arthur-Conti/guh/packages/log/logger"

type BaseConfigs struct {
	Logger *logger.Logger
}

var Config = &BaseConfigs{}

func Init() {
	Config.Logger = InitLogger()
}