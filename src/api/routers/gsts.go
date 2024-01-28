package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jaganathanb/dapps-api/api/handlers"
	"github.com/jaganathanb/dapps-api/api/middlewares"
	"github.com/jaganathanb/dapps-api/config"
)

func Gst(router *gin.RouterGroup, cfg *config.Config) {
	h := handlers.NewGstsHandler(cfg)

	router.POST("/", middlewares.Authorization([]string{"admin"}), h.CreateGsts)
	router.POST("/page", middlewares.Authorization([]string{"admin"}), h.GetGsts)
	router.GET("/", middlewares.Authorization([]string{"admin"}), h.GetGsts)
	router.PUT("/:gstin/statuses", middlewares.Authorization([]string{"admin"}), h.UpdateGstStatuses)
}
