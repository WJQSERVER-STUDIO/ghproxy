package auth

import (
	"encoding/json"
	"ghproxy/config"
	"os"
)

type WhitelistConfig struct {
	Whitelist []string `json:"whitelist"`
}

var (
	whitelistfile = "/data/ghproxy/config/whitelist.json"
	whitelist     *WhitelistConfig
)

func LoadWhitelist(cfg *config.Config) {
	whitelistfile = cfg.Whitelist.WhitelistFile
	whitelist = &WhitelistConfig{}

	data, err := os.ReadFile(whitelistfile)
	if err != nil {
		logw("Failed to read whitelist file: %v", err)
	}

	err = json.Unmarshal(data, whitelist)
	if err != nil {
		logw("Failed to unmarshal whitelist JSON: %v", err)
	}
}

func CheckWhitelist(fullrepo string) bool {
	return forRangeCheckWhitelist(whitelist.Whitelist, fullrepo)
}

func forRangeCheckWhitelist(blist []string, fullrepo string) bool {
	for _, blocked := range blist {
		if blocked == fullrepo {
			return true
		}
	}
	return false
}
