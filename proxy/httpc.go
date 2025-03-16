package proxy

import (
	"fmt"
	"ghproxy/config"
	"net/http"
	"sync"
	"time"

	httpc "github.com/satomitouka/touka-httpc"
)

var BufferSize int = 32 * 1024 // 32KB

var (
	tr         *http.Transport
	gittr      *http.Transport
	BufferPool *sync.Pool
	client     *httpc.Client
	gitclient  *httpc.Client
)

func InitReq(cfg *config.Config) {
	initHTTPClient(cfg)
	if cfg.GitClone.Mode == "cache" {
		initGitHTTPClient(cfg)
	}

	// 初始化固定大小的缓存池
	BufferPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, BufferSize)
		},
	}
}

func initHTTPClient(cfg *config.Config) {
	var proTolcols = new(http.Protocols)
	proTolcols.SetHTTP1(true)
	proTolcols.SetHTTP2(true)
	proTolcols.SetUnencryptedHTTP2(true)
	if cfg.Httpc.Mode == "auto" {

		tr = &http.Transport{
			//MaxIdleConns:    160,
			IdleConnTimeout: 30 * time.Second,
			WriteBufferSize: 32 * 1024, // 32KB
			ReadBufferSize:  32 * 1024, // 32KB
			Protocols:       proTolcols,
		}
	} else if cfg.Httpc.Mode == "advanced" {
		tr = &http.Transport{
			MaxIdleConns:        cfg.Httpc.MaxIdleConns,
			MaxConnsPerHost:     cfg.Httpc.MaxConnsPerHost,
			MaxIdleConnsPerHost: cfg.Httpc.MaxIdleConnsPerHost,
			WriteBufferSize:     32 * 1024, // 32KB
			ReadBufferSize:      32 * 1024, // 32KB
			Protocols:           proTolcols,
		}
	} else {
		// 错误的模式
		logError("unknown httpc mode: %s", cfg.Httpc.Mode)
		fmt.Println("unknown httpc mode: ", cfg.Httpc.Mode)
		logWarning("use Auto to Run HTTP Client")
		fmt.Println("use Auto to Run HTTP Client")
		tr = &http.Transport{
			//MaxIdleConns:    160,
			IdleConnTimeout: 30 * time.Second,
			WriteBufferSize: 32 * 1024, // 32KB
			ReadBufferSize:  32 * 1024, // 32KB
		}
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
}

func initGitHTTPClient(cfg *config.Config) {

	var proTolcols = new(http.Protocols)
	proTolcols.SetHTTP1(true)
	proTolcols.SetHTTP2(true)
	proTolcols.SetUnencryptedHTTP2(true)
	if cfg.GitClone.ForceH2C {
		proTolcols.SetHTTP1(false)
		proTolcols.SetHTTP2(false)
		proTolcols.SetUnencryptedHTTP2(true)
	}
	if cfg.Httpc.Mode == "auto" {

		gittr = &http.Transport{
			//MaxIdleConns:    160,
			IdleConnTimeout: 30 * time.Second,
			WriteBufferSize: 32 * 1024, // 32KB
			ReadBufferSize:  32 * 1024, // 32KB
			Protocols:       proTolcols,
		}
	} else if cfg.Httpc.Mode == "advanced" {
		gittr = &http.Transport{
			MaxIdleConns:        cfg.Httpc.MaxIdleConns,
			MaxConnsPerHost:     cfg.Httpc.MaxConnsPerHost,
			MaxIdleConnsPerHost: cfg.Httpc.MaxIdleConnsPerHost,
			WriteBufferSize:     32 * 1024, // 32KB
			ReadBufferSize:      32 * 1024, // 32KB
			Protocols:           proTolcols,
		}
	} else {
		// 错误的模式
		logError("unknown httpc mode: %s", cfg.Httpc.Mode)
		fmt.Println("unknown httpc mode: ", cfg.Httpc.Mode)
		logWarning("use Auto to Run HTTP Client")
		fmt.Println("use Auto to Run HTTP Client")
		gittr = &http.Transport{
			//MaxIdleConns:    160,
			IdleConnTimeout: 30 * time.Second,
			WriteBufferSize: 32 * 1024, // 32KB
			ReadBufferSize:  32 * 1024, // 32KB
		}
	}
	if cfg.Outbound.Enabled {
		initTransport(cfg, gittr)
	}
	if cfg.Server.Debug {
		gitclient = httpc.New(
			httpc.WithTransport(gittr),
			httpc.WithDumpLog(),
		)
	} else {
		gitclient = httpc.New(
			httpc.WithTransport(gittr),
		)
	}
}
