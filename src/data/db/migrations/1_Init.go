package migrations

import (
	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/constants"
	"github.com/jaganathanb/dapps-api/data/db"
	"github.com/jaganathanb/dapps-api/data/models"
	"github.com/jaganathanb/dapps-api/pkg/logging"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var logger = logging.NewLogger(config.GetConfig())

func Up_1(cfg *config.Config) {
	database := db.GetDb()

	createTables(database)
	createDefaultUserInformation(database, cfg)
	createOrUpdateSettings(database, cfg)

	//actual migrations
	alterColumns(database)
}

func alterColumns(db *gorm.DB) {
	err := db.Migrator().AddColumn(&models.Gst{}, "Sno")
	if err != nil {
		logger.Error(logging.Sqlite3, logging.Migration, err.Error(), nil)
	} else {
		logger.Info(logging.Sqlite3, logging.Migration, "Column Sno created", nil)
	}

	err = db.Migrator().AddColumn(&models.Gst{}, "Fno")
	if err != nil {
		logger.Error(logging.Sqlite3, logging.Migration, err.Error(), nil)
	} else {
		logger.Info(logging.Sqlite3, logging.Migration, "Column Fno created", nil)
	}

	err = db.Migrator().AddColumn(&models.Gst{}, "Username")
	if err != nil {
		logger.Error(logging.Sqlite3, logging.Migration, err.Error(), nil)
	} else {
		logger.Info(logging.Sqlite3, logging.Migration, "Column Username created", nil)
	}

	err = db.Migrator().AddColumn(&models.Gst{}, "Password")
	if err != nil {
		logger.Error(logging.Sqlite3, logging.Migration, err.Error(), nil)
	} else {
		logger.Info(logging.Sqlite3, logging.Migration, "Column Password created", nil)
	}
}

func createTables(database *gorm.DB) {
	tables := []interface{}{}

	// User
	tables = addNewTable(database, models.User{}, tables)
	tables = addNewTable(database, models.Role{}, tables)
	tables = addNewTable(database, models.UserRole{}, tables)
	tables = addNewTable(database, models.Gst{}, tables)
	tables = addNewTable(database, models.AdditionalAddress{}, tables)
	tables = addNewTable(database, models.Address{}, tables)
	tables = addNewTable(database, models.PermenantAddress{}, tables)
	tables = addNewTable(database, models.GstStatus{}, tables)
	tables = addNewTable(database, models.Settings{}, tables)
	tables = addNewTable(database, models.Notifications{}, tables)

	err := database.Migrator().CreateTable(tables...)
	if err != nil {
		logger.Error(logging.Postgres, logging.Migration, err.Error(), nil)
	}
	logger.Info(logging.Postgres, logging.Migration, "tables created", nil)
}

func addNewTable(database *gorm.DB, model interface{}, tables []interface{}) []interface{} {
	if !database.Migrator().HasTable(model) {
		tables = append(tables, model)
	}

	return tables
}

func createDefaultUserInformation(database *gorm.DB, cfg *config.Config) {

	adminRole := models.Role{Name: constants.AdminRoleName}
	createRoleIfNotExists(database, &adminRole)

	defaultRole := models.Role{Name: constants.DefaultRoleName}
	createRoleIfNotExists(database, &defaultRole)

	u := models.User{Username: cfg.Server.Username, FirstName: "", LastName: "",
		MobileNumber: "09111112222", Email: cfg.Server.Username}
	pass := cfg.Server.Password

	logger.Infof("Verifying user %s, if exists, it will not create.", u.Username)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	u.Password = string(hashedPassword)

	createAdminUserIfNotExists(database, &u, adminRole.Id)
}

func createOrUpdateSettings(database *gorm.DB, cfg *config.Config) {
	settings := models.Settings{}

	err := database.
		Model(&models.Settings{}).
		Where("1 = 1").
		FirstOrInit(&settings).Error

	if err == nil {
		if cfg.Server.Gst.Username != "" {
			settings.GstUsername = cfg.Server.Gst.Username
		} else {
			cfg.Server.Gst.Username = settings.GstUsername
		}

		if cfg.Server.Gst.Password != "" {
			settings.GstPassword = cfg.Server.Gst.Password
		} else {
			cfg.Server.Gst.Password = settings.GstPassword
		}

		if cfg.Server.Gst.BaseUrl != "" {
			settings.GstBaseUrl = cfg.Server.Gst.BaseUrl
		} else {
			cfg.Server.Gst.BaseUrl = settings.GstBaseUrl
		}

		if cfg.Server.Gst.Crontab != "" {
			settings.Crontab = cfg.Server.Gst.Crontab
		} else {
			cfg.Server.Gst.Crontab = settings.Crontab
		}

		if settings.Id == 0 {
			database.Create(&settings)
		} else {
			database.Updates(&settings)
		}
	}
}

func createRoleIfNotExists(database *gorm.DB, r *models.Role) {
	exists := 0
	database.
		Model(&models.Role{}).
		Select("1").
		Where("name = ?", r.Name).
		First(&exists)
	if exists == 0 {
		database.Create(r)
	}
}

func createAdminUserIfNotExists(database *gorm.DB, u *models.User, roleId int) {
	exists := 0
	database.
		Model(&models.User{}).
		Select("1").
		Where("username = ?", u.Username).
		First(&exists)
	if exists == 0 {
		database.Create(u)
		ur := models.UserRole{UserId: u.Id, RoleId: roleId}
		database.Create(&ur)
	}
}

func Down_1() {

}
