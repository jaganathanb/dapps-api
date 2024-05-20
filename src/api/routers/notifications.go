package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jaganathanb/dapps-api/api/handlers"
	"github.com/jaganathanb/dapps-api/config"
)

func Notifications(r *gin.RouterGroup, cfg *config.Config) {
	handler := handlers.NewNotificationsHandler(cfg)

	r.PUT("", handler.UpdateNotifications)
	r.POST("", handler.AddNotifications)
	r.GET("", handler.GetNotifications)
	r.DELETE("", handler.DeleteNotifications)
}
