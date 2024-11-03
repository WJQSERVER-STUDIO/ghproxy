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

// fullrepo: "owner/repo" or "owner/*"
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
	// 先匹配user,再匹配user/*,最后匹配完整repo
	for _, blocked := range blist {
		// 切片
		users, _ := sliceRepoName_Blacklist(blocked)
		logw("users:%s, blocked:%s", users, blocked)
		// 匹配 user
		if user == users {
			// 匹配 user/*
			if strings.HasSuffix(blocked, "/*") {
				return true
			}
			// 匹配完整repo
			if fullrepo == blocked {
				return true
			}
		}
	}

	/*	for _, blocked := range blist {
		if blocked == fullrepo || (strings.HasSuffix(blocked, "/*") && strings.HasPrefix(repoUser, blocked[:len(blocked)-2])) {
			return true
		}
	} */
	return false
}
