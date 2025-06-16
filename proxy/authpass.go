package proxy

import (
	"ghproxy/config"
	"net/http"

	"github.com/infinite-iroha/touka"
)

func AuthPassThrough(c *touka.Context, cfg *config.Config, req *http.Request) {
	if cfg.Auth.PassThrough {
		token := c.Query("token")
		if token != "" {
			switch cfg.Auth.Method {
			case "parameters":
				if !cfg.Auth.Enabled {
					req.Header.Set("Authorization", "token "+token)
				} else {
					c.Warnf("%s %s %s %s %s Auth-Error: Conflict Auth Method", c.ClientIP(), c.Request.Method, c.Request.URL.Path, c.UserAgent(), c.Request.Proto)
					ErrorPage(c, NewErrorWithStatusLookup(500, "Conflict Auth Method"))
					return
				}
			case "header":
				if cfg.Auth.Enabled {
					req.Header.Set("Authorization", "token "+token)
				}
			default:
				c.Warnf("%s %s %s %s %s Invalid Auth Method / Auth Method is not be set", c.ClientIP(), c.Request.Method, c.Request.URL.Path, c.UserAgent(), c.Request.Proto)
				ErrorPage(c, NewErrorWithStatusLookup(500, "Invalid Auth Method / Auth Method is not be set"))
				return
			}
		}
	}
}
