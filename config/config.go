package config

import (
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

type Config struct {
	CachePath   string `yaml:"cachePath"`
	InstallPath string `yaml:"installPath"`
}

func InitConfig(processDir, configPath string) (conf Config, err error) {
	allBytes, err := os.ReadFile(configPath)
	if err != nil {
		conf = Config{
			CachePath:   filepath.Join(processDir, ".cache"),
			InstallPath: filepath.Join(processDir, ".install"),
		}
		allBytes, err = yaml.Marshal(conf)
		os.WriteFile(configPath, allBytes, 0777)
		return
	}
	err = yaml.Unmarshal(allBytes, &conf)
	return
}
