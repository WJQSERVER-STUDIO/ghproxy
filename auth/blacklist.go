package auth

import (
	"encoding/json"
	"fmt"
	"ghproxy/config"
	"os"
	"strings"
	"sync"
)

type Blacklist struct {
	userSet     map[string]struct{}            // 用户级黑名单
	repoSet     map[string]map[string]struct{} // 仓库级黑名单
	initOnce    sync.Once                      // 确保初始化只执行一次
	initialized bool                           // 初始化状态标识
}

var (
	instance *Blacklist
	initErr  error
)

// InitBlacklist 初始化黑名单（线程安全，仅执行一次）
func InitBlacklist(cfg *config.Config) error {
	instance = &Blacklist{
		userSet: make(map[string]struct{}),
		repoSet: make(map[string]map[string]struct{}),
	}

	data, err := os.ReadFile(cfg.Blacklist.BlacklistFile)
	if err != nil {
		return fmt.Errorf("failed to read blacklist: %w", err)
	}

	var list struct {
		Entries []string `json:"blacklist"`
	}
	if err := json.Unmarshal(data, &list); err != nil {
		return fmt.Errorf("invalid blacklist format: %w", err)
	}

	for _, entry := range list.Entries {
		user, repo := splitUserRepo(entry)
		switch {
		case repo == "" || repo == "*":
			instance.userSet[user] = struct{}{}
		default:
			if _, exists := instance.repoSet[user]; !exists {
				instance.repoSet[user] = make(map[string]struct{})
			}
			instance.repoSet[user][repo] = struct{}{}
		}
	}

	instance.initialized = true
	return nil
}

// CheckBlacklist 检查用户和仓库是否在黑名单中（无锁设计）
func CheckBlacklist(username, repo string) bool {
	if instance == nil || !instance.initialized {
		return false
	}

	// 先检查用户级黑名单
	if _, exists := instance.userSet[username]; exists {
		return true
	}

	// 再检查仓库级黑名单
	if repos, userExists := instance.repoSet[username]; userExists {
		// 允许仓库名为空时的全用户仓库匹配
		if repo == "" {
			return true
		}
		_, repoExists := repos[repo]
		return repoExists
	}

	return false
}

// splitUserRepo 优化分割逻辑（仅初始化时使用）
func splitUserRepo(fullRepo string) (user, repo string) {
	if idx := strings.Index(fullRepo, "/"); idx > 0 {
		return fullRepo[:idx], fullRepo[idx+1:]
	}
	return fullRepo, ""
}
