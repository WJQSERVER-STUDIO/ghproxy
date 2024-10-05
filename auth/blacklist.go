package auth

import (
	"encoding/json"
	"ghproxy/config"
	"os"
)

type Config struct {
	Blacklist []string `json:"blacklist"`
}

var (
	cfg           *config.Config
	blacklistfile = "/data/ghproxy/config/blacklist.json"
	blacklist     *Config
)

func LoadBlacklist(cfg *config.Config) {
	blacklistfile = cfg.Blacklist.BlacklistFile
	blacklist = &Config{}

	data, err := os.ReadFile(blacklistfile)
	if err != nil {
		logw("Failed to read blacklist file: %v", err)
	}

	err = json.Unmarshal(data, blacklist)
	if err != nil {
		logw("Failed to unmarshal blacklist JSON: %v", err)
	}
}

func CheckBlacklist(fullrepo string) bool {
	return forRangeCheck(blacklist.Blacklist, fullrepo)
}

func forRangeCheck(blist []string, fullrepo string) bool {
	for _, blocked := range blist {
		if blocked == fullrepo {
			return true
		}
	}
	logw("%s not in blacklist", fullrepo)
	return false
}
