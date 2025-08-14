package envlocations

import (
	"os"

	errorhandler "github.com/Arthur-Conti/guh/libs/error_handler"
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
		return errorhandler.Wrap(errorhandler.KindInternal, "error loading "+le.FilePath, err, errorhandler.WithOp("env_locations.LoadDotEnv"), errorhandler.WithFields(map[string]any{"filePath": le.FilePath}))
	}
	return nil
}

func (le *LocalEnvs) Get(key string) string {
	val := os.Getenv(key)
	return val
}

func (le *LocalEnvs) GetOrDefault(key, defaultVal string) string {
	val := le.Get(key)
	if val == "" {
		return defaultVal
	}
	return val
}
