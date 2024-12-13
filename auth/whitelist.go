package auth

import (
	"encoding/json"
	"github.com/satomitoka/ghproxy/config"
	"os"
	"strings"
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
		logError("Failed to read whitelist file: %v", err)
	}

	err = json.Unmarshal(data, whitelist)
	if err != nil {
		logError("Failed to unmarshal whitelist JSON: %v", err)
	}
}

func CheckWhitelist(fullrepo string, user string, repo string) bool {
	return forRangeCheckWhitelist(whitelist.Whitelist, fullrepo, user)
}

func sliceRepoName_Whitelist(fullrepo string) (string, string) {
	s := strings.Split(fullrepo, "/")
	if len(s) != 2 {
		return "", ""
	}
	return s[0], s[1]
}

func forRangeCheckWhitelist(wlist []string, fullrepo string, user string) bool {
	for _, passd := range wlist {
		users, _ := sliceRepoName_Whitelist(passd)
		if users == user {
			if strings.HasSuffix(passd, "/*") {
				return true
			}
			if fullrepo == passd {
				return true
			}
		}
	}
	return false
}
