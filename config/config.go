package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	CachePath   string `yaml:"cachePath"`
	InstallPath string `yaml:"installPath"`
	AutoSetEnv  *bool  `yaml:"autoSetEnv"` // 自动设置环境变量
	path        string
}

func (c *Config) Sync() {
	allBytes, _ := yaml.Marshal(c)
	_ = os.WriteFile(c.path, allBytes, 0o777)
}

func InitConfig(processDir, configPath string) (conf Config, err error) {
	allBytes, err := os.ReadFile(configPath)
	if err != nil {
		conf = Config{
			CachePath:   filepath.Join(processDir, ".cache"),
			InstallPath: filepath.Join(processDir, ".install"),
			path:        configPath,
		}
		allBytes, _ = yaml.Marshal(conf)
		err = os.WriteFile(configPath, allBytes, 0o777)
		return
	}
	err = yaml.Unmarshal(allBytes, &conf)
	conf.path = configPath
	return
}
