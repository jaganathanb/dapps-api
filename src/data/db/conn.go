package db

import (
	"fmt"
	"log"
	"path"
	"time"

	"github.com/jaganathanb/dapps-api/config"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var dbClient *gorm.DB

func InitDb(cfg *config.Config) error {
	var err error
	switch cfg.Server.DB {
	case "postgres":
		cnn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Tehran",
			cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password,
			cfg.Postgres.DbName, cfg.Postgres.SSLMode)

		dbClient, err = gorm.Open(postgres.Open(cnn), &gorm.Config{})
		if err != nil {
			return err
		}

		sqlDb, _ := dbClient.DB()
		err = sqlDb.Ping()
		if err != nil {
			return err
		}

		sqlDb.SetMaxIdleConns(cfg.Postgres.MaxIdleConns)
		sqlDb.SetMaxOpenConns(cfg.Postgres.MaxOpenConns)
		sqlDb.SetConnMaxLifetime(cfg.Postgres.ConnMaxLifetime * time.Minute)

	default:
		cnn := path.Join(cfg.Sqlite3.DbName, "dapps_gst.db")

		dbClient, err = gorm.Open(sqlite.Open(cnn), &gorm.Config{})
		if err != nil {
			return err
		}

		sqlDb, _ := dbClient.DB()
		err = sqlDb.Ping()
		if err != nil {
			return err
		}
	}

	log.Println("Db connection established")
	return nil
}

func GetDb() *gorm.DB {
	return dbClient
}

func CloseDb() {
	con, _ := dbClient.DB()
	con.Close()
}
