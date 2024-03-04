package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jaganathanb/dapps-api/api/handlers"
	"github.com/jaganathanb/dapps-api/api/middlewares"
	"github.com/jaganathanb/dapps-api/config"
)

func Gst(router *gin.RouterGroup, cfg *config.Config) {
	h := handlers.NewGstsHandler(cfg)

	if cfg.Server.RunMode == "release" {
		router.Use(middlewares.Authentication(cfg), middlewares.Authorization([]string{"admin"}))
	}

	router.POST("/", h.CreateGsts)
	router.POST("/page", h.GetGsts)
	router.GET("/", h.GetGsts)
	router.PUT("/:gstin/statuses", h.UpdateGstStatuses)
	router.PUT("/:gstin/lock", h.LockGstById)
}
