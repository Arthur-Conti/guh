package projectconfig

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

const configFilePath string = ".guh.yaml"

type ProjectConfig struct {
	ServiceName  string `yaml:"serviceName"`
	ModName      string `yaml:"modName"`
	BaseUrl      string `yaml:"baseUrl"`
	DbUser       string `yaml:"dbUser"`
	DbPassword   string `yaml:"dbPassword"`
	DbIP         string `yaml:"dbIP"`
	DbPort       string `yaml:"dbPort"`
	DbDatabase   string `yaml:"dbDatabase"`
}

func Load() (*ProjectConfig, error) {
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &ProjectConfig{}, nil
		}
		return nil, err
	}
	cfg := &ProjectConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func Save(cfg *ProjectConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(configFilePath, data, 0644)
}
