package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port      int    `yaml:"port"`
		Host      string `yaml:"host"`
		SizeLimit int    `yaml:"sizelimit"`
	} `yaml:"server"`

	Log struct {
		LogFilePath string `yaml:"logfilepath"`
		MaxLogSize  int    `yaml:"maxlogsize"`
	} `yaml:"logger"`

	CORS struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"cors"`

	Auth struct {
		Enabled   bool   `yaml:"enabled"`
		AuthToken string `yaml:"authtoken"`
	} `yaml:"auth"`

	Blacklist struct {
		Enabled       bool   `yaml:"enabled"`
		BlacklistFile string `yaml:"blacklistfile"`
	} `yaml:"blacklist"`
}

type Blacklist struct {
	Blacklist map[string][]string `yaml:"blacklist"`
}

// LoadConfig 从 YAML 配置文件加载配置
func LoadConfig(filePath string) (*Config, error) {
	var config Config
	if err := loadYAML(filePath, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// LoadBlacklistConfig 从 YAML 配置文件加载黑名单配置
func LoadBlacklistConfig(filePath string) (*Blacklist, error) {
	var config Blacklist
	if err := loadYAML(filePath, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// LoadyamlConfig 从 YAML 配置文件加载配置
func loadYAML(filePath string, out interface{}) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, out)
}
