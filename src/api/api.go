package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/jaganathanb/dapps-api/api/middlewares"
	"github.com/jaganathanb/dapps-api/api/routers"
	validation "github.com/jaganathanb/dapps-api/api/validations"
	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/docs"
	"github.com/jaganathanb/dapps-api/pkg/logging"
	"github.com/jaganathanb/dapps-api/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var logger = logging.NewLogger(config.GetConfig())

func InitServer(cfg *config.Config) {
	gin.SetMode(cfg.Server.RunMode)
	r := gin.New()

	RegisterValidators()
	RegisterPrometheus()

	r.Use(middlewares.DefaultStructuredLogger(cfg))
	r.Use(middlewares.Cors(cfg))
	r.Use(middlewares.Prometheus())
	r.Use(gin.Logger(), gin.CustomRecovery(middlewares.ErrorHandler) /*middlewares.TestMiddleware()*/, middlewares.LimitByRequest())

	RegisterRoutes(r, cfg)
	RegisterSwagger(r, cfg)

	logger := logging.NewLogger(cfg)
	logger.Info(logging.General, logging.Startup, "Started", nil)
	err := r.Run(fmt.Sprintf(":%s", cfg.Server.InternalPort))
	if err != nil {
		logger.Fatal(logging.General, logging.Startup, err.Error(), nil)
	}
}

func RegisterRoutes(r *gin.Engine, cfg *config.Config) {
	api := r.Group("/api")

	v1 := api.Group("/v1")
	{
		// Test
		health := v1.Group("/health")
		test_router := v1.Group("/test" /*middlewares.Authentication(cfg), middlewares.Authorization([]string{"admin"})*/)

		// User, GSTs
		auth := v1.Group("/auth")
		gsts := v1.Group("/gsts")
		gst := gsts.Group("/:gstin")

		mocks := v1.Group("/mocks")

		streamer := v1.Group("/stream")

		// Test
		routers.Health(health)
		routers.TestRouter(test_router)

		// User, GST, Mock & Streamer
		routers.Auth(auth, cfg)
		routers.Gsts(gsts, cfg)
		routers.Gst(gst, cfg)
		routers.Mock(mocks, cfg)
		routers.Streamer(streamer, cfg)

		r.Static("/static", "./uploads")

		r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	v2 := api.Group("/v2")
	{
		health := v2.Group("/health")
		routers.Health(health)
	}
}

func RegisterValidators() {
	val, ok := binding.Validator.Engine().(*validator.Validate)
	if ok {
		err := val.RegisterValidation("mobile", validation.IndianMobileNumberValidator, true)
		if err != nil {
			logger.Error(logging.Validation, logging.Startup, err.Error(), nil)
		}
		err = val.RegisterValidation("password", validation.PasswordValidator, true)
		if err != nil {
			logger.Error(logging.Validation, logging.Startup, err.Error(), nil)
		}
		err = val.RegisterValidation("gstins", validation.GstinsValidator, true)
		if err != nil {
			logger.Error(logging.Validation, logging.Startup, err.Error(), nil)
		}
		err = val.RegisterValidation("gstin", validation.GstinValidator, true)
		if err != nil {
			logger.Error(logging.Validation, logging.Startup, err.Error(), nil)
		}
	}
}

func RegisterSwagger(r *gin.Engine, cfg *config.Config) {
	docs.SwaggerInfo.Title = "DApps Web API"
	docs.SwaggerInfo.Description = "ReST API for DApps portal"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/api"
	docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%s", cfg.Server.ExternalPort)
	docs.SwaggerInfo.Schemes = []string{"http"}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func RegisterPrometheus() {
	err := prometheus.Register(metrics.DbCall)
	if err != nil {
		logger.Error(logging.Prometheus, logging.Startup, err.Error(), nil)
	}

	err = prometheus.Register(metrics.HttpDuration)
	if err != nil {
		logger.Error(logging.Prometheus, logging.Startup, err.Error(), nil)
	}
}
