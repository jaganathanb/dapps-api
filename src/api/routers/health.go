package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jaganathanb/dapps-api/api/handlers"
)

func Health(r *gin.RouterGroup) {
	handler := handlers.NewHealthHandler()

	r.GET("/", handler.Health)
}
