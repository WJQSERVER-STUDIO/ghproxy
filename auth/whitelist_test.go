package auth

import (
	"os"
	"path/filepath"
	"testing"

	"ghproxy/config"
)

func TestSplitUserRepoWhitelist_PreservesHyphenatedOwner(t *testing.T) {
	user, repo := splitUserRepoWhitelist("jow-/ucode")

	if user != "jow-" {
		t.Fatalf("user = %q, want %q", user, "jow-")
	}
	if repo != "ucode" {
		t.Fatalf("repo = %q, want %q", repo, "ucode")
	}
}

func TestInitWhitelist_AllowsHyphenatedOwnerRepo(t *testing.T) {
	tmpDir := t.TempDir()
	whitelistPath := filepath.Join(tmpDir, "whitelist.json")

	content := []byte(`{"whitelist":["jow-/ucode"]}`)
	if err := os.WriteFile(whitelistPath, content, 0o600); err != nil {
		t.Fatalf("write whitelist file: %v", err)
	}

	whitelistInstance = nil
	whitelistInitErr = nil

	cfg := &config.Config{}
	cfg.Whitelist.WhitelistFile = whitelistPath

	if err := InitWhitelist(cfg); err != nil {
		t.Fatalf("InitWhitelist() error = %v", err)
	}

	if !CheckWhitelist("jow-", "ucode") {
		t.Fatal("CheckWhitelist() = false, want true for jow-/ucode")
	}

	if CheckWhitelist("jow", "ucode") {
		t.Fatal("CheckWhitelist() = true, want false for non-hyphenated owner")
	}

	if CheckWhitelist("jow-", "other") {
		t.Fatal("CheckWhitelist() = true, want false for unmatched repo")
	}

	t.Cleanup(func() {
		whitelistInstance = nil
		whitelistInitErr = nil
	})
}
