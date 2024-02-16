package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/jaganathanb/dapps-api/config"
)

func CustomHeaders(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Writer.Header().Set("apiKey", cfg.Server.GstApiKey)

		c.Next()
	}
}
