package proxy

import (
	"ghproxy/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthPassThrough(c *gin.Context, cfg *config.Config, req *http.Request) {
	if cfg.Auth.PassThrough {
		token := c.Query("token")
		if token != "" {
			logDebug("%s %s %s %s %s Auth-PassThrough: token %s", c.ClientIP(), c.Request.Method, c.Request.URL.String(), c.Request.Header.Get("User-Agent"), c.Request.Proto, token)
			switch cfg.Auth.AuthMethod {
			case "parameters":
				if !cfg.Auth.Enabled {
					req.Header.Set("Authorization", "token "+token)
				} else {
					logWarning("%s %s %s %s %s Auth-Error: Conflict Auth Method", c.ClientIP(), c.Request.Method, c.Request.URL.String(), c.Request.Header.Get("User-Agent"), c.Request.Proto)
					// 500 Internal Server Error
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Conflict Auth Method"})
					return
				}
			case "header":
				if cfg.Auth.Enabled {
					req.Header.Set("Authorization", "token "+token)
				}
			default:
				logWarning("%s %s %s %s %s Invalid Auth Method / Auth Method is not be set", c.ClientIP(), c.Request.Method, c.Request.URL.String(), c.Request.Header.Get("User-Agent"), c.Request.Proto)
				// 500 Internal Server Error
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid Auth Method / Auth Method is not be set"})
				return
			}
		}
	}
}
