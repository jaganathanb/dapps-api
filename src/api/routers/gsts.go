package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jaganathanb/dapps-api/api/handlers"
	"github.com/jaganathanb/dapps-api/api/middlewares"
	"github.com/jaganathanb/dapps-api/config"
)

func Gst(router *gin.RouterGroup, cfg *config.Config) {
	h := handlers.NewGstsHandler(cfg)

	router.POST("/", middlewares.Authentication(cfg), middlewares.Authorization([]string{"admin"}), h.CreateGST)
}
