package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jaganathanb/dapps-api/api/handlers"
	"github.com/jaganathanb/dapps-api/config"
)

func Settings(r *gin.RouterGroup, cfg *config.Config) {
	handler := handlers.NewSettingsHandler(cfg)

	r.GET("", handler.GetSettings)
	r.PUT("", handler.UpdateSettings)
}
