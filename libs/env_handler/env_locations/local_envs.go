package envlocations

import (
	"os"

	"github.com/Arthur-Conti/guh/config"
	errorhandler "github.com/Arthur-Conti/guh/libs/error_handler"
	"github.com/Arthur-Conti/guh/libs/log/logger"
	"github.com/joho/godotenv"
)

type LocalEnvs struct {
	FilePath string
}

func NewLocalEnvs(filePath string) *LocalEnvs {
	return &LocalEnvs{FilePath: filePath}
}

func (le *LocalEnvs) LoadDotEnv() error {
	if err := godotenv.Load(le.FilePath); err != nil {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "envlocations", Message: "Error loading %v: %v\n", Vals: []any{le.FilePath, err}})
		return errorhandler.Wrap("InternalServerError", "error loading "+le.FilePath, err)
	}
	return nil
}

func (le *LocalEnvs) Get(key string) string {
	val := os.Getenv(key)
	if val == "" {
		config.Config.Logger.Errorf(logger.LogMessage{ApplicationPackage: "envlocations", Message: "Env key %v not found", Vals: []any{key}})
	}
	return val
}

func (le *LocalEnvs) GetOrDefault(key, defaultVal string) string {
	val := le.Get(key)
	if val == "" {
		return defaultVal
	}
	return val
}
