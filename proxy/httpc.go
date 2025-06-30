package proxy

import (
	"ghproxy/config"
	"net/http"
	"time"

	"github.com/WJQSERVER-STUDIO/httpc"
)

var BufferSize int = 32 * 1024 // 32KB

var (
	tr        *http.Transport
	gittr     *http.Transport
	client    *httpc.Client
	gitclient *httpc.Client
)

func InitReq(cfg *config.Config) (*httpc.Client, error) {
	client := initHTTPClient(cfg)
	if cfg.GitClone.Mode == "cache" {
		initGitHTTPClient(cfg)
	}
	err := SetGlobalRateLimit(cfg)
	if err != nil {
		return nil, err
	}
	return client, nil

}

func initHTTPClient(cfg *config.Config) *httpc.Client {
	var proTolcols = new(http.Protocols)
	proTolcols.SetHTTP1(true)
	proTolcols.SetHTTP2(true)
	proTolcols.SetUnencryptedHTTP2(true)

	switch cfg.Httpc.Mode {
	case "auto", "":
		tr = &http.Transport{
			IdleConnTimeout: 30 * time.Second,
			WriteBufferSize: 32 * 1024, // 32KB
			ReadBufferSize:  32 * 1024, // 32KB
			Protocols:       proTolcols,
		}
	case "advanced":
		tr = &http.Transport{
			MaxIdleConns:        cfg.Httpc.MaxIdleConns,
			MaxConnsPerHost:     cfg.Httpc.MaxConnsPerHost,
			MaxIdleConnsPerHost: cfg.Httpc.MaxIdleConnsPerHost,
			WriteBufferSize:     32 * 1024, // 32KB
			ReadBufferSize:      32 * 1024, // 32KB
			Protocols:           proTolcols,
		}
	default:
		panic("unknown httpc mode: " + cfg.Httpc.Mode)
	}

	if cfg.Outbound.Enabled {
		initTransport(cfg, tr)
	}
	if cfg.Server.Debug {
		client = httpc.New(
			httpc.WithTransport(tr),
			httpc.WithDumpLog(),
		)
	} else {
		client = httpc.New(
			httpc.WithTransport(tr),
		)
	}
	return client
}

func initGitHTTPClient(cfg *config.Config) {
	switch cfg.Httpc.Mode {
	case "auto", "":
		gittr = &http.Transport{
			IdleConnTimeout: 30 * time.Second,
			WriteBufferSize: 32 * 1024, // 32KB
			ReadBufferSize:  32 * 1024, // 32KB
		}
	case "advanced":
		gittr = &http.Transport{
			MaxIdleConns:        cfg.Httpc.MaxIdleConns,
			MaxConnsPerHost:     cfg.Httpc.MaxConnsPerHost,
			MaxIdleConnsPerHost: cfg.Httpc.MaxIdleConnsPerHost,
			WriteBufferSize:     32 * 1024, // 32KB
			ReadBufferSize:      32 * 1024, // 32KB
		}
	default:
		panic("unknown httpc mode: " + cfg.Httpc.Mode)
	}

	if cfg.Outbound.Enabled {
		initTransport(cfg, gittr)
	}

	var opts []httpc.Option // 使用切片来收集选项
	opts = append(opts, httpc.WithTransport(gittr))
	var protocolsConfig httpc.ProtocolsConfig

	if cfg.GitClone.ForceH2C {
		protocolsConfig.ForceH2C = true
	} else {
		protocolsConfig.Http1 = true
		protocolsConfig.Http2 = true
		protocolsConfig.Http2_Cleartext = true
	}
	opts = append(opts, httpc.WithProtocols(protocolsConfig))

	if cfg.Server.Debug {
		opts = append(opts, httpc.WithDumpLog())
	}

	gitclient = httpc.New(opts...)
}
