package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/WJQSERVER/wanf"
)

// Config 结构体定义了整个应用程序的配置
type Config struct {
	Server    ServerConfig    `toml:"server" wanf:"server"`
	Httpc     HttpcConfig     `toml:"httpc" wanf:"httpc"`
	GitClone  GitCloneConfig  `toml:"gitclone" wanf:"gitclone"`
	Shell     ShellConfig     `toml:"shell" wanf:"shell"`
	Pages     PagesConfig     `toml:"pages" wanf:"pages"`
	Log       LogConfig       `toml:"log" wanf:"log"`
	Auth      AuthConfig      `toml:"auth" wanf:"auth"`
	Blacklist BlacklistConfig `toml:"blacklist" wanf:"blacklist"`
	Whitelist WhitelistConfig `toml:"whitelist" wanf:"whitelist"`
	IPFilter  IPFilterConfig  `toml:"ipFilter" wanf:"ipFilter"`
	RateLimit RateLimitConfig `toml:"rateLimit" wanf:"rateLimit"`
	Outbound  OutboundConfig  `toml:"outbound" wanf:"outbound"`
	Docker    DockerConfig    `toml:"docker" wanf:"docker"`
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

// ServerConfig 定义服务器相关的配置
type ServerConfig struct {
	Port      int    `toml:"port" wanf:"port"`
	Host      string `toml:"host" wanf:"host"`
	SizeLimit int    `toml:"sizeLimit" wanf:"sizeLimit"`
	MemLimit  int64  `toml:"memLimit" wanf:"memLimit"`
	Cors      string `toml:"cors" wanf:"cors"`
	Debug     bool   `toml:"debug" wanf:"debug"`
}

/*
[httpc]
mode = "auto" # "auto" or "advanced"
maxIdleConns = 100 # only for advanced mode
maxIdleConnsPerHost = 60 # only for advanced mode
maxConnsPerHost = 0 # only for advanced mode
useCustomRawHeaders = false
*/
// HttpcConfig 定义 HTTP 客户端相关的配置
type HttpcConfig struct {
	Mode                string `toml:"mode" wanf:"mode"`
	MaxIdleConns        int    `toml:"maxIdleConns" wanf:"maxIdleConns"`
	MaxIdleConnsPerHost int    `toml:"maxIdleConnsPerHost" wanf:"maxIdleConnsPerHost"`
	MaxConnsPerHost     int    `toml:"maxConnsPerHost" wanf:"maxConnsPerHost"`
	UseCustomRawHeaders bool   `toml:"useCustomRawHeaders" wanf:"useCustomRawHeaders"`
}

/*
[gitclone]
mode = "bypass" # bypass / cache
smartGitAddr = "http://127.0.0.1:8080"
//cacheTimeout = 10
ForceH2C = true
*/
// GitCloneConfig 定义 Git 克隆相关的配置
type GitCloneConfig struct {
	Mode         string `toml:"mode" wanf:"mode"`
	SmartGitAddr string `toml:"smartGitAddr" wanf:"smartGitAddr"`
	//CacheTimeout int    `toml:"cacheTimeout"`
	ForceH2C bool `toml:"ForceH2C" wanf:"ForceH2C"`
}

/*
[shell]
editor = true
rewriteAPI = false
*/
// ShellConfig 定义 Shell 相关的配置
type ShellConfig struct {
	Editor     bool `toml:"editor" wanf:"editor"`
	RewriteAPI bool `toml:"rewriteAPI" wanf:"rewriteAPI"`
}

/*
[pages]
mode = "internal" # "internal" or "external"
theme = "bootstrap" # "bootstrap" or "nebula"
staticDir = "/data/www"
*/
// PagesConfig 定义静态页面相关的配置
type PagesConfig struct {
	Mode      string `toml:"mode" wanf:"mode"`
	Theme     string `toml:"theme" wanf:"theme"`
	StaticDir string `toml:"staticDir" wanf:"staticDir"`
}

// LogConfig 定义日志相关的配置
type LogConfig struct {
	LogFilePath string `toml:"logFilePath" wanf:"logFilePath"`
	MaxLogSize  int64  `toml:"maxLogSize" wanf:"maxLogSize"`
	Level       string `toml:"level" wanf:"level"`
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
// AuthConfig 定义认证相关的配置
type AuthConfig struct {
	Enabled               bool   `toml:"enabled" wanf:"enabled"`
	Method                string `toml:"method" wanf:"method"`
	Key                   string `toml:"key" wanf:"key"`
	Token                 string `toml:"token" wanf:"token"`
	PassThrough           bool   `toml:"passThrough" wanf:"passThrough"`
	ForceAllowApi         bool   `toml:"ForceAllowApi" wanf:"ForceAllowApi"`
	ForceAllowApiPassList bool   `toml:"ForceAllowApiPassList" wanf:"ForceAllowApiPassList"`
}

// BlacklistConfig 定义黑名单相关的配置
type BlacklistConfig struct {
	Enabled       bool   `toml:"enabled" wanf:"enabled"`
	BlacklistFile string `toml:"blacklistFile" wanf:"blacklistFile"`
}

// WhitelistConfig 定义白名单相关的配置
type WhitelistConfig struct {
	Enabled       bool   `toml:"enabled" wanf:"enabled"`
	WhitelistFile string `toml:"whitelistFile" wanf:"whitelistFile"`
}

// IPFilterConfig 定义 IP 过滤相关的配置
type IPFilterConfig struct {
	Enabled         bool   `toml:"enabled" wanf:"enabled"`
	EnableAllowList bool   `toml:"enableAllowList" wanf:"enableAllowList"`
	EnableBlockList bool   `toml:"enableBlockList" wanf:"enableBlockList"`
	IPFilterFile    string `toml:"ipFilterFile" wanf:"ipFilterFile"`
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

// RateLimitConfig 定义限速相关的配置
type RateLimitConfig struct {
	Enabled        bool                 `toml:"enabled" wanf:"enabled"`
	RatePerMinute  int                  `toml:"ratePerMinute" wanf:"ratePerMinute"`
	Burst          int                  `toml:"burst" wanf:"burst"`
	BandwidthLimit BandwidthLimitConfig `toml:"bandwidthLimit" wanf:"bandwidthLimit"`
}

// BandwidthLimitConfig 定义带宽限制相关的配置
type BandwidthLimitConfig struct {
	Enabled     bool   `toml:"enabled" wanf:"enabled"`
	TotalLimit  string `toml:"totalLimit" wanf:"totalLimit"`
	TotalBurst  string `toml:"totalBurst" wanf:"totalBurst"`
	SingleLimit string `toml:"singleLimit" wanf:"singleLimit"`
	SingleBurst string `toml:"singleBurst" wanf:"singleBurst"`
}

/*
[outbound]
enabled = false
url = "socks5://127.0.0.1:1080" # "http://127.0.0.1:7890"
*/
// OutboundConfig 定义出站代理相关的配置
type OutboundConfig struct {
	Enabled bool   `toml:"enabled" wanf:"enabled"`
	Url     string `toml:"url" wanf:"url"`
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
// DockerConfig 定义 Docker 相关的配置
type DockerConfig struct {
	Enabled         bool              `toml:"enabled" wanf:"enabled"`
	Target          string            `toml:"target" wanf:"target"`
	Auth            bool              `toml:"auth" wanf:"auth"`
	Credentials     map[string]string `toml:"credentials" wanf:"credentials"`
	AuthPassThrough bool              `toml:"authPassThrough" wanf:"authPassThrough"`
}

// LoadConfig 从配置文件加载配置
func LoadConfig(filePath string) (*Config, error) {
	exist, filePath2read := FileExists(filePath)
	if !exist {
		// 楔入配置文件
		err := DefaultConfig().WriteConfig(filePath)
		if err != nil {
			return nil, err
		}
		return DefaultConfig(), nil
	}
	var config Config
	ext := filepath.Ext(filePath2read)
	if ext == ".wanf" {
		if err := wanf.DecodeFile(filePath2read, &config); err != nil {
			return nil, err
		}
		return &config, nil
	}

	if _, err := toml.DecodeFile(filePath2read, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// WriteConfig 写入配置文件
func (c *Config) WriteConfig(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	ext := filepath.Ext(filePath)
	if ext == ".wanf" {
		err := wanf.NewStreamEncoder(file).Encode(c)
		if err != nil {
			return err
		}
		return nil
	}

	encoder := toml.NewEncoder(file)
	return encoder.Encode(c)
}

// FileExists 检测文件是否存在
func FileExists(filename string) (bool, string) {
	_, err := os.Stat(filename)
	if err == nil {
		return true, filename
	}
	if os.IsNotExist(err) {
		// 获取文件名（不包含路径）
		base := filepath.Base(filename)
		dir := filepath.Dir(filename)

		// 获取扩展名
		fileNameBody := strings.TrimSuffix(base, filepath.Ext(base))

		// 重新组合路径, 扩展名改为.wanf, 确认是否存在
		wanfFilename := filepath.Join(dir, fileNameBody+".wanf")

		_, err = os.Stat(wanfFilename)
		if err == nil {
			// .wanf 文件存在
			fmt.Printf("\n Found .wanf file: %s\n", wanfFilename)
			return true, wanfFilename
		} else if os.IsNotExist(err) {
			// .wanf 文件不存在
			return false, ""
		} else {
			// 其他错误
			return false, ""
		}
	} else {
		return false, filename
	}
}

// DefaultConfig 返回默认配置结构体
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
