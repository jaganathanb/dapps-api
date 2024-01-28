package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jaganathanb/dapps-api/api/handlers"
	"github.com/jaganathanb/dapps-api/config"
)

func Mock(r *gin.RouterGroup, cfg *config.Config) {
	handler := handlers.NewMockHandler(cfg)

	r.GET("/:filename/:prop", handler.GetMockData)
}
