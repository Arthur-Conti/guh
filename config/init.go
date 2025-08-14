package config

import (
	envhandler "github.com/Arthur-Conti/guh/libs/env_handler"
	"github.com/Arthur-Conti/guh/libs/log/logger"
)

type BaseConfigs struct {
	Logger *logger.Logger
	Env    *envhandler.Envs
}

var Config = &BaseConfigs{}

func Init() {
	Config.Logger = InitLogger()
	Config.Env = InitEnv()
}
