package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jaganathanb/dapps-api/api/handlers"
	"github.com/jaganathanb/dapps-api/api/middlewares"
	"github.com/jaganathanb/dapps-api/config"
)

func User(router *gin.RouterGroup, cfg *config.Config) {
	h := handlers.NewAuthHandler(cfg)

	if cfg.Server.RunMode == "release" {
		router.Use(middlewares.Authentication(cfg), middlewares.Authorization([]string{"admin", "default"}))
	}

	router.GET("/profile", h.GetLoggedInUserDetail)
}
