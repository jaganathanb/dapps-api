package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jaganathanb/dapps-api/api/handlers"
	"github.com/jaganathanb/dapps-api/api/middlewares"
	"github.com/jaganathanb/dapps-api/config"
)

func Streamer(router *gin.RouterGroup, cfg *config.Config) {
	handler := handlers.NewStreamerHandler(cfg)

	if cfg.Server.RunMode == "release" {
		router.Use(middlewares.AuthenticationByQueryString(cfg), middlewares.Authorization([]string{"admin", "default"}))
	}

	router.GET("/", middlewares.StreamerHeaders(cfg), handler.ServeStream(), handler.StreamData)
}
