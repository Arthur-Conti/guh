package envhandler

import envlocations "github.com/Arthur-Conti/guh/libs/env_handler/env_locations"

type Envs struct {
	EnvLocation envlocations.EnvLocationsInterface
}

func NewEnvs(location envlocations.EnvLocationsInterface) *Envs {
	return &Envs{EnvLocation: location}
}
