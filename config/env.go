package config

import (
	envhandler "github.com/Arthur-Conti/guh/libs/env_handler"
	envlocations "github.com/Arthur-Conti/guh/libs/env_handler/env_locations"
)

func InitEnv() *envhandler.Envs {
	loc := envlocations.NewLocalEnvs("./.env")
	return envhandler.NewEnvs(loc)
}