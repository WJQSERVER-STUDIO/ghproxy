package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server    ServerConfig    `toml:"server"`
	Httpc     HttpcConfig     `toml:"httpc"`
	GitClone  GitCloneConfig  `toml:"gitclone"`
	Shell     ShellConfig     `toml:"shell"`
	Pages     PagesConfig     `toml:"pages"`
	Log       LogConfig       `toml:"log"`
	Auth      AuthConfig      `toml:"auth"`
	Blacklist BlacklistConfig `toml:"blacklist"`
	Whitelist WhitelistConfig `toml:"whitelist"`
	IPFilter  IPFilterConfig  `toml:"ipFilter"`
	RateLimit RateLimitConfig `toml:"rateLimit"`
	Outbound  OutboundConfig  `toml:"outbound"`
	Docker    DockerConfig    `toml:"docker"`
}

/*
[server]
host = "0.0.0.0"
port = 8080
sizeLimit = 125 # MB
memLimit = 0 # MB
cors = "*" # "*"/"" -> "*" ; "nil" -> "" ;
debug = false
*/

type ServerConfig struct {
	Port      int    `toml:"port"`
	Host      string `toml:"host"`
	SizeLimit int    `toml:"sizeLimit"`
	MemLimit  int64  `toml:"memLimit"`
	Cors      string `toml:"cors"`
	Debug     bool   `toml:"debug"`
}

/*
[httpc]
mode = "auto" # "auto" or "advanced"
maxIdleConns = 100 # only for advanced mode
maxIdleConnsPerHost = 60 # only for advanced mode
maxConnsPerHost = 0 # only for advanced mode
useCustomRawHeaders = false
*/
type HttpcConfig struct {
	Mode                string `toml:"mode"`
	MaxIdleConns        int    `toml:"maxIdleConns"`
	MaxIdleConnsPerHost int    `toml:"maxIdleConnsPerHost"`
	MaxConnsPerHost     int    `toml:"maxConnsPerHost"`
	UseCustomRawHeaders bool   `toml:"useCustomRawHeaders"`
}

/*
[gitclone]
mode = "bypass" # bypass / cache
smartGitAddr = "http://127.0.0.1:8080"
//cacheTimeout = 10
ForceH2C = true
*/
type GitCloneConfig struct {
	Mode         string `toml:"mode"`
	SmartGitAddr string `toml:"smartGitAddr"`
	//CacheTimeout int    `toml:"cacheTimeout"`
	ForceH2C bool `toml:"ForceH2C"`
}

/*
[shell]
editor = true
rewriteAPI = false
*/
type ShellConfig struct {
	Editor     bool `toml:"editor"`
	RewriteAPI bool `toml:"rewriteAPI"`
}

/*
[pages]
mode = "internal" # "internal" or "external"
theme = "bootstrap" # "bootstrap" or "nebula"
staticDir = "/data/www"
*/
type PagesConfig struct {
	Mode      string `toml:"mode"`
	Theme     string `toml:"theme"`
	StaticDir string `toml:"staticDir"`
}

type LogConfig struct {
	LogFilePath string `toml:"logFilePath"`
	MaxLogSize  int64  `toml:"maxLogSize"`
	Level       string `toml:"level"`
}

/*
[auth]
Method = "parameters" # "header" or "parameters"
Key = ""
Token = "token"
enabled = false
passThrough = false
ForceAllowApi = false
ForceAllowApiPassList = false
*/
type AuthConfig struct {
	Enabled               bool   `toml:"enabled"`
	Method                string `toml:"method"`
	Key                   string `toml:"key"`
	Token                 string `toml:"token"`
	PassThrough           bool   `toml:"passThrough"`
	ForceAllowApi         bool   `toml:"ForceAllowApi"`
	ForceAllowApiPassList bool   `toml:"ForceAllowApiPassList"`
}

type BlacklistConfig struct {
	Enabled       bool   `toml:"enabled"`
	BlacklistFile string `toml:"blacklistFile"`
}

type WhitelistConfig struct {
	Enabled       bool   `toml:"enabled"`
	WhitelistFile string `toml:"whitelistFile"`
}

type IPFilterConfig struct {
	Enabled         bool   `toml:"enabled"`
	EnableAllowList bool   `toml:"enableAllowList"`
	EnableBlockList bool   `toml:"enableBlockList"`
	IPFilterFile    string `toml:"ipFilterFile"`
}

/*
[rateLimit]
enabled = false
ratePerMinute = 100
burst = 10

	[rateLimit.bandwidthLimit]
	enabled = false
	totalLimit = "100mbps"
	totalBurst = "100mbps"
	singleLimit = "10mbps"
	singleBurst = "10mbps"
*/

type RateLimitConfig struct {
	Enabled        bool `toml:"enabled"`
	RatePerMinute  int  `toml:"ratePerMinute"`
	Burst          int  `toml:"burst"`
	BandwidthLimit BandwidthLimitConfig
}

type BandwidthLimitConfig struct {
	Enabled     bool   `toml:"enabled"`
	TotalLimit  string `toml:"totalLimit"`
	TotalBurst  string `toml:"totalBurst"`
	SingleLimit string `toml:"singleLimit"`
	SingleBurst string `toml:"singleBurst"`
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

/*
[docker]
enabled = false
target = "ghcr" # ghcr/dockerhub
auth = false
[docker.credentials]
user1 = "testpass"
test = "test123"
*/
type DockerConfig struct {
	Enabled         bool              `toml:"enabled"`
	Target          string            `toml:"target"`
	Auth            bool              `toml:"auth"`
	Credentials     map[string]string `toml:"credentials"`
	AuthPassThrough bool              `toml:"authPassThrough"`
}

// LoadConfig 从 TOML 配置文件加载配置
func LoadConfig(filePath string) (*Config, error) {
	if !FileExists(filePath) {
		// 楔入配置文件
		err := DefaultConfig().WriteConfig(filePath)
		if err != nil {
			return nil, err
		}
		return DefaultConfig(), nil
	}

	var config Config
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// 写入配置文件
func (c *Config) WriteConfig(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	return encoder.Encode(c)
}

// 检测文件是否存在
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// 默认配置结构体
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:      8080,
			Host:      "0.0.0.0",
			SizeLimit: 125,
			MemLimit:  0,
			Cors:      "*",
			Debug:     false,
		},
		Httpc: HttpcConfig{
			Mode:                "auto",
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 60,
			MaxConnsPerHost:     0,
		},
		GitClone: GitCloneConfig{
			Mode:         "bypass",
			SmartGitAddr: "http://127.0.0.1:8080",
			ForceH2C:     false,
		},
		Shell: ShellConfig{
			Editor:     false,
			RewriteAPI: false,
		},
		Pages: PagesConfig{
			Mode:      "internal",
			Theme:     "hub",
			StaticDir: "/data/www",
		},
		Log: LogConfig{
			LogFilePath: "/data/ghproxy/log/ghproxy.log",
			MaxLogSize:  10,
			Level:       "info",
		},
		Auth: AuthConfig{
			Enabled:               false,
			Method:                "parameters",
			Key:                   "",
			Token:                 "token",
			PassThrough:           false,
			ForceAllowApi:         false,
			ForceAllowApiPassList: false,
		},
		Blacklist: BlacklistConfig{
			Enabled:       false,
			BlacklistFile: "/data/ghproxy/config/blacklist.json",
		},
		Whitelist: WhitelistConfig{
			Enabled:       false,
			WhitelistFile: "/data/ghproxy/config/whitelist.json",
		},
		IPFilter: IPFilterConfig{
			Enabled:         false,
			IPFilterFile:    "/data/ghproxy/config/ipfilter.json",
			EnableAllowList: false,
			EnableBlockList: false,
		},
		RateLimit: RateLimitConfig{
			Enabled:       false,
			RatePerMinute: 100,
			Burst:         10,
			BandwidthLimit: BandwidthLimitConfig{
				Enabled:     false,
				TotalLimit:  "100mbps",
				TotalBurst:  "100mbps",
				SingleLimit: "10mbps",
				SingleBurst: "10mbps",
			},
		},
		Outbound: OutboundConfig{
			Enabled: false,
			Url:     "socks5://127.0.0.1:1080",
		},
		Docker: DockerConfig{
			Enabled: false,
			Target:  "dockerhub",
			Auth:    false,
			Credentials: map[string]string{
				"testpass": "test123",
			},
		},
	}
}
