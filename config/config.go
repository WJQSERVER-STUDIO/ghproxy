package config

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Server    ServerConfig
	Log       LogConfig
	CORS      CORSConfig
	Auth      AuthConfig
	Blacklist BlacklistConfig
	Whitelist WhitelistConfig
}

type ServerConfig struct {
	Port      int    `toml:"port"`
	Host      string `toml:"host"`
	SizeLimit int    `toml:"sizelimit"`
}

type LogConfig struct {
	LogFilePath string `toml:"logfilepath"`
	MaxLogSize  int    `toml:"maxlogsize"`
}

type CORSConfig struct {
	Enabled bool `toml:"enabled"`
}

type AuthConfig struct {
	Enabled   bool   `toml:"enabled"`
	AuthToken string `toml:"authtoken"`
}

type BlacklistConfig struct {
	Enabled       bool   `toml:"enabled"`
	BlacklistFile string `toml:"blacklistfile"`
}

type WhitelistConfig struct {
	Enabled       bool   `toml:"enabled"`
	WhitelistFile string `toml:"whitelistfile"`
}

// LoadConfig 从 TOML 配置文件加载配置
func LoadConfig(filePath string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
