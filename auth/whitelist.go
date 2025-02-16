package auth

import (
	"encoding/json"
	"fmt"
	"ghproxy/config"
	"os"
	"strings"
	"sync"
)

// Whitelist 用于存储白名单信息
type Whitelist struct {
	userSet     map[string]struct{}            // 用户级白名单
	repoSet     map[string]map[string]struct{} // 仓库级白名单
	initOnce    sync.Once                      // 确保初始化只执行一次
	initialized bool                           // 初始化状态标识
}

var (
	whitelistInstance *Whitelist
	whitelistInitErr  error
)

// InitWhitelist 初始化白名单（线程安全，仅执行一次）
func InitWhitelist(cfg *config.Config) error {
	whitelistInstance = &Whitelist{
		userSet: make(map[string]struct{}),
		repoSet: make(map[string]map[string]struct{}),
	}

	data, err := os.ReadFile(cfg.Whitelist.WhitelistFile)
	if err != nil {
		return fmt.Errorf("failed to read whitelist: %w", err)
	}

	var list struct {
		Entries []string `json:"whitelist"`
	}
	if err := json.Unmarshal(data, &list); err != nil {
		return fmt.Errorf("invalid whitelist format: %w", err)
	}

	for _, entry := range list.Entries {
		user, repo := splitUserRepoWhitelist(entry)
		switch {
		case repo == "" || repo == "*":
			whitelistInstance.userSet[user] = struct{}{}
		default:
			if _, exists := whitelistInstance.repoSet[user]; !exists {
				whitelistInstance.repoSet[user] = make(map[string]struct{})
			}
			whitelistInstance.repoSet[user][repo] = struct{}{}
		}
	}

	whitelistInstance.initialized = true
	return nil
}

// CheckWhitelist 检查用户和仓库是否在白名单中（无锁设计）
func CheckWhitelist(username, repo string) bool {
	if whitelistInstance == nil || !whitelistInstance.initialized {
		return false
	}

	// 先检查用户级白名单
	if _, exists := whitelistInstance.userSet[username]; exists {
		return true
	}

	// 再检查仓库级白名单
	if repos, userExists := whitelistInstance.repoSet[username]; userExists {
		// 允许仓库名为空时的全用户仓库匹配
		if repo == "" {
			return true
		}
		_, repoExists := repos[repo]
		return repoExists
	}

	return false
}

// splitUserRepoWhitelist 分割用户和仓库信息（仅初始化时使用）
func splitUserRepoWhitelist(fullRepo string) (user, repo string) {
	if idx := strings.Index(fullRepo, "/"); idx > 0 {
		return fullRepo[:idx], fullRepo[idx+1:]
	}
	return fullRepo, ""
}
