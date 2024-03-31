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

func StreamerHeaders(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", cfg.Cors.AllowOrigins)
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "GET")
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Transfer-Encoding", "chunked")

		c.Next()
	}
}
