package proxy

import (
	"ghproxy/config"
	"testing"
)

func TestInitHTTPClient_AutoModeUsesConfiguredPoolSizes(t *testing.T) {
	oldTr, oldClient := tr, client
	t.Cleanup(func() {
		tr = oldTr
		client = oldClient
	})

	cfg := &config.Config{}
	cfg.Httpc.Mode = "auto"
	cfg.Httpc.MaxIdleConns = 123
	cfg.Httpc.MaxIdleConnsPerHost = 45
	cfg.Httpc.MaxConnsPerHost = 67

	initHTTPClient(cfg)

	if tr == nil {
		t.Fatal("transport was not initialized")
	}
	if tr.MaxIdleConns != 123 {
		t.Fatalf("MaxIdleConns = %d, want 123", tr.MaxIdleConns)
	}
	if tr.MaxIdleConnsPerHost != 45 {
		t.Fatalf("MaxIdleConnsPerHost = %d, want 45", tr.MaxIdleConnsPerHost)
	}
	if tr.MaxConnsPerHost != 67 {
		t.Fatalf("MaxConnsPerHost = %d, want 67", tr.MaxConnsPerHost)
	}
}

func TestInitGitHTTPClient_AutoModeUsesConfiguredPoolSizes(t *testing.T) {
	oldGitTr, oldGitClient := gittr, gitclient
	t.Cleanup(func() {
		gittr = oldGitTr
		gitclient = oldGitClient
	})

	cfg := &config.Config{}
	cfg.Httpc.Mode = "auto"
	cfg.Httpc.MaxIdleConns = 98
	cfg.Httpc.MaxIdleConnsPerHost = 76
	cfg.Httpc.MaxConnsPerHost = 54

	initGitHTTPClient(cfg)

	if gittr == nil {
		t.Fatal("git transport was not initialized")
	}
	if gittr.MaxIdleConns != 98 {
		t.Fatalf("MaxIdleConns = %d, want 98", gittr.MaxIdleConns)
	}
	if gittr.MaxIdleConnsPerHost != 76 {
		t.Fatalf("MaxIdleConnsPerHost = %d, want 76", gittr.MaxIdleConnsPerHost)
	}
	if gittr.MaxConnsPerHost != 54 {
		t.Fatalf("MaxConnsPerHost = %d, want 54", gittr.MaxConnsPerHost)
	}
}
