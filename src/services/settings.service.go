package services

import (
	"database/sql"
	"sync"

	"github.com/jaganathanb/dapps-api/api/dto"
	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/data/db"
	"github.com/jaganathanb/dapps-api/data/models"
	"github.com/jaganathanb/dapps-api/pkg/logging"
	scrap_scheduler "github.com/jaganathanb/dapps-api/pkg/scrap-scheduler"
	"gorm.io/gorm"
)

type SettingsService struct {
	logger     logging.Logger
	cfg        *config.Config
	database   *gorm.DB
	gstService GstService
	scheduler  scrap_scheduler.DAppsJobScheduler
}

var settingsService *SettingsService
var settingsServiceOnce sync.Once

func NewSettingsService(cfg *config.Config) *SettingsService {
	settingsServiceOnce.Do(func() {
		database := db.GetDb()
		logger := logging.NewLogger(cfg)
		settingsService = &SettingsService{
			cfg:        cfg,
			database:   database,
			logger:     logger,
			gstService: *NewGstService(cfg),
			scheduler:  *scrap_scheduler.NewDAppsJobScheduler(cfg),
		}
	})

	return settingsService
}

// Get settings
func (s *SettingsService) GetSettings() (*dto.SettingsPayload, error) {
	settings := models.Settings{}

	result := s.database.
		Model(&models.Settings{}).
		Find(&settings)

	if result.Error != nil {
		return nil, result.Error
	}

	return &dto.SettingsPayload{
		Crontab:     settings.Crontab,
		GstUsername: settings.GstUsername,
		GstPassword: settings.GstPassword,
		GstBaseUrl:  settings.GstBaseUrl,
		BaseDto: dto.BaseDto{
			Id: settings.BaseModel.Id,
		},
	}, nil
}

// Get settings
func (s *SettingsService) UpdateSettings(req *dto.SettingsPayload) (*dto.SettingsPayload, error) {
	var settings models.Settings

	s.database.Model(&models.Settings{}).First(&settings)

	crontab := settings.Crontab

	tx := s.database.Begin()

	settings.Id = req.Id
	settings.ModifiedBy = &sql.NullInt64{
		Valid: true,
		Int64: int64(req.ModifiedBy),
	}
	settings.CreatedBy = req.ModifiedBy
	settings.Crontab = req.Crontab
	settings.GstUsername = req.GstUsername
	settings.GstPassword = req.GstPassword
	settings.GstBaseUrl = req.GstBaseUrl

	err := tx.Model(&models.Settings{}).Where("id = ?", settings.Id).Save(&settings).Error

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if crontab != req.Crontab {
		s.scheduler.RemoveJobs("gst")
		s.scheduler.AddJob(settings.Crontab, "gst", s.gstService.scrapGstPortal, req.ModifiedBy)
	}

	tx.Commit()

	s.cfg.Server.Gst.Username = settings.GstUsername
	s.cfg.Server.Gst.Password = settings.GstPassword
	s.cfg.Server.Gst.Crontab = settings.Crontab

	return &dto.SettingsPayload{
		Crontab:     settings.Crontab,
		GstUsername: settings.GstUsername,
		GstPassword: settings.GstPassword,
		GstBaseUrl:  settings.GstBaseUrl,
		BaseDto: dto.BaseDto{
			Id: settings.BaseModel.Id,
		},
	}, nil
}
