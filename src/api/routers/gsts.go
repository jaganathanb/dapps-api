package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jaganathanb/dapps-api/api/handlers"
	"github.com/jaganathanb/dapps-api/api/middlewares"
	"github.com/jaganathanb/dapps-api/config"
)

func Gsts(router *gin.RouterGroup, cfg *config.Config) {
	h := handlers.NewGstsHandler(cfg)

	if cfg.Server.RunMode == "release" {
		router.Use(middlewares.Authentication(cfg), middlewares.Authorization([]string{"admin", "default"}))
	}

	router.POST("/create", h.CreateGsts)
	router.POST("/page", h.GetGsts)
	router.GET("/statistics", h.GetGstStatistics)
	router.GET("/refresh-returns", h.RefreshGstReturns)
}

func Gst(router *gin.RouterGroup, cfg *config.Config) {
	h := handlers.NewGstsHandler(cfg)

	if cfg.Server.RunMode == "release" {
		router.Use(middlewares.Authentication(cfg), middlewares.Authorization([]string{"admin", "default"}))
	}

	router.PUT("/return-status", h.UpdateGstStatus)
	router.PUT("/lock", h.LockGstById)
}
