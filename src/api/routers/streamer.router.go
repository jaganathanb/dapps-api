package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jaganathanb/dapps-api/api/handlers"
	"github.com/jaganathanb/dapps-api/config"
)

func Streamer(r *gin.RouterGroup, cfg *config.Config) {
	handler := handlers.NewStreamerHandler(cfg)

	r.GET("/", handler.StreamData)
}
