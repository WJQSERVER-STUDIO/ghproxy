package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port        int    `yaml:"port"`
	Host        string `yaml:"host"`
	SizeLimit   int    `yaml:"sizelimit"`
	LogFilePath string `yaml:"logfilepath"`
	CORSOrigin  bool   `yaml:"CorsAllowOrigins"`
	Auth        bool   `yaml:"auth"`
	AuthToken   string `yaml:"authtoken"`
}

// LoadConfig 从 YAML 配置文件加载配置
func LoadConfig(filePath string) (*Config, error) {
	var config Config
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
