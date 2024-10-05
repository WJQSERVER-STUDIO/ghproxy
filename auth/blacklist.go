package auth

import (
	"fmt"
	"ghproxy/config"
	"log"
)

var (
	cfg        *config.Config
	configfile = "/data/ghproxy/config/config.yaml"
	blacklist  *config.Blist
)

func init() {
	loadBlacklistConfig()
}

func loadConfig() {
	var err error
	// 初始化配置
	cfg, err = config.LoadConfig(configfile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("Loaded config: %v\n", cfg)
}

func loadBlacklistConfig() {
	var err error
	// 初始化黑名单配置
	blacklist, err = config.LoadBlacklistConfig(cfg.Blacklist.BlacklistFile)
	if err != nil {
		log.Fatalf("Failed to load blacklist: %v", err)
	}
	logw("Loaded blacklist: %v", blacklist)
}

func CheckBlacklist(fullrepo string) bool {
	forRangeCheck(blacklist.Blacklist, fullrepo)
	return false
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
