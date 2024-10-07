package auth

import (
	"encoding/json"
	"ghproxy/config"
	"os"
)

type BlacklistConfig struct {
	Blacklist []string `json:"blacklist"`
}

var (
	cfg           *config.Config
	blacklistfile = "/data/ghproxy/config/blacklist.json"
	blacklist     *BlacklistConfig
)

func LoadBlacklist(cfg *config.Config) {
	blacklistfile = cfg.Blacklist.BlacklistFile
	blacklist = &BlacklistConfig{}

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
	return forRangeCheckBlacklist(blacklist.Blacklist, fullrepo)
}

func forRangeCheckBlacklist(blist []string, fullrepo string) bool {
	for _, blocked := range blist {
		if blocked == fullrepo {
			return true
		}
	}
	return false
}
