package main

import (
	"github.com/jaganathanb/dapps-api/api"
	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/data/db"
	"github.com/jaganathanb/dapps-api/data/db/migrations"
	"github.com/jaganathanb/dapps-api/pkg/logging"
)

// @securityDefinitions.apikey AuthBearer
// @in header
// @name Authorization
func main() {
	cfg := config.GetConfig()

	logger := logging.NewLogger(cfg)

	// err := cache.InitRedis(cfg)
	// defer cache.CloseRedis()
	// if err != nil {
	// 	logger.Fatal(logging.Redis, logging.Startup, err.Error(), nil)
	// }

	err := db.InitDb(cfg)
	defer db.CloseDb()
	if err != nil {
		logger.Fatal(logging.Postgres, logging.Startup, err.Error(), nil)
	}
	migrations.Up_1(cfg)

	api.InitServer(cfg)
}
