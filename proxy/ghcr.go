package proxy

import (
	"context"
	"ghproxy/config"

	"github.com/cloudwego/hertz/pkg/app"
)

func GhcrRouting(cfg *config.Config) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		ChunkedProxyRequest(ctx, c, "https://ghcr.io"+string(c.Request.RequestURI()), cfg, "ghcr")
	}
}
