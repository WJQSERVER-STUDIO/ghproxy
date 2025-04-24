package proxy

import (
	"ghproxy/config"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

func AuthPassThrough(c *app.RequestContext, cfg *config.Config, req *http.Request) {
	if cfg.Auth.PassThrough {
		token := c.Query("token")
		if token != "" {
			logDebug("%s %s %s %s %s Auth-PassThrough: token %s", c.ClientIP(), c.Method(), string(c.Path()), c.UserAgent(), c.Request.Header.GetProtocol(), token)
			switch cfg.Auth.Method {
			case "parameters":
				if !cfg.Auth.Enabled {
					req.Header.Set("Authorization", "token "+token)
				} else {
					logWarning("%s %s %s %s %s Auth-Error: Conflict Auth Method", c.ClientIP(), c.Method(), string(c.Path()), c.UserAgent(), c.Request.Header.GetProtocol())
					ErrorPage(c, NewErrorWithStatusLookup(500, "Conflict Auth Method"))
					return
				}
			case "header":
				if cfg.Auth.Enabled {
					req.Header.Set("Authorization", "token "+token)
				}
			default:
				logWarning("%s %s %s %s %s Invalid Auth Method / Auth Method is not be set", c.ClientIP(), c.Method(), string(c.Path()), c.UserAgent(), c.Request.Header.GetProtocol())
				ErrorPage(c, NewErrorWithStatusLookup(500, "Invalid Auth Method / Auth Method is not be set"))
				return
			}
		}
	}
}
