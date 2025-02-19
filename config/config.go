package config

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Server    ServerConfig
	Httpc     HttpcConfig
	Pages     PagesConfig
	Log       LogConfig
	CORS      CORSConfig
	Auth      AuthConfig
	Blacklist BlacklistConfig
	Whitelist WhitelistConfig
	RateLimit RateLimitConfig
	Outbound  OutboundConfig
}

type ServerConfig struct {
	Port      int    `toml:"port"`
	Host      string `toml:"host"`
	SizeLimit int    `toml:"sizeLimit"`
	EnableH2C string `toml:"enableH2C"`
	Debug     bool   `toml:"debug"`
}

/*
[httpc]
mode = "auto" # "auto" or "advanced"
maxIdleConns = 100 # only for advanced mode
maxIdleConnsPerHost = 60 # only for advanced mode
maxConnsPerHost = 0 # only for advanced mode
*/
type HttpcConfig struct {
	Mode                string `toml:"mode"`
	MaxIdleConns        int    `toml:"maxIdleConns"`
	MaxIdleConnsPerHost int    `toml:"maxIdleConnsPerHost"`
	MaxConnsPerHost     int    `toml:"maxConnsPerHost"`
}

/*
[pages]
enabled = false
theme = "bootstrap" # "bootstrap" or "nebula"
staticDir = "/data/www"
*/
type PagesConfig struct {
	Enabled   bool   `toml:"enabled"`
	Theme     string `toml:"theme"`
	StaticDir string `toml:"staticDir"`
}

type LogConfig struct {
	LogFilePath string `toml:"logFilePath"`
	MaxLogSize  int    `toml:"maxLogSize"`
	Level       string `toml:"level"`
}

type CORSConfig struct {
	Enabled bool `toml:"enabled"`
}

type AuthConfig struct {
	Enabled     bool   `toml:"enabled"`
	AuthMethod  string `toml:"authMethod"`
	AuthToken   string `toml:"authToken"`
	PassThrough bool   `toml:"passThrough"`
}

type BlacklistConfig struct {
	Enabled       bool   `toml:"enabled"`
	BlacklistFile string `toml:"blacklistFile"`
}

type WhitelistConfig struct {
	Enabled       bool   `toml:"enabled"`
	WhitelistFile string `toml:"whitelistFile"`
}

type RateLimitConfig struct {
	Enabled       bool   `toml:"enabled"`
	RateMethod    string `toml:"rateMethod"`
	RatePerMinute int    `toml:"ratePerMinute"`
	Burst         int    `toml:"burst"`
}

/*
[outbound]
enabled = false
url = "socks5://127.0.0.1:1080" # "http://127.0.0.1:7890"
*/
type OutboundConfig struct {
	Enabled bool   `toml:"enabled"`
	Url     string `toml:"url"`
}

// LoadConfig 从 TOML 配置文件加载配置
func LoadConfig(filePath string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
