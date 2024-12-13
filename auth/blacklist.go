package auth

import (
	"encoding/json"
	"ghproxy/config"
	"os"
	"strings"
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
		logError("Failed to read blacklist file: %v", err)
	}

	err = json.Unmarshal(data, blacklist)
	if err != nil {
		logError("Failed to unmarshal blacklist JSON: %v", err)
	}
}

func CheckBlacklist(repouser string, user string, repo string) bool {
	return forRangeCheckBlacklist(blacklist.Blacklist, repouser, user)
}

func sliceRepoName_Blacklist(fullrepo string) (string, string) {
	s := strings.Split(fullrepo, "/")
	if len(s) != 2 {
		return "", ""
	}
	return s[0], s[1]
}

func forRangeCheckBlacklist(blist []string, fullrepo string, user string) bool {
	for _, blocked := range blist {
		users, _ := sliceRepoName_Blacklist(blocked)
		if user == users {
			if strings.HasSuffix(blocked, "/*") {
				return true
			}
			if fullrepo == blocked {
				return true
			}
		}
	}
	return false
}
