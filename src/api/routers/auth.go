package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jaganathanb/dapps-api/api/handlers"
	"github.com/jaganathanb/dapps-api/api/middlewares"
	"github.com/jaganathanb/dapps-api/config"
)

func Auth(router *gin.RouterGroup, cfg *config.Config) {
	h := handlers.NewAuthHandler(cfg)

	router.POST("/send-otp", middlewares.OtpLimiter(cfg), middlewares.Authentication(cfg), middlewares.Authorization([]string{"admin"}), h.SendOtp)
	router.POST("/login", h.LoginByUsername)
	router.POST("/logout", h.LogoutByUsername)
	router.POST("/register", h.RegisterByUsername)
	router.POST("/login-m", h.RegisterLoginByMobileNumber)
}
