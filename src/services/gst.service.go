package services

import (
	"github.com/jaganathanb/dapps-api/api/dto"
	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/data/db"
	"github.com/jaganathanb/dapps-api/data/models"
	"github.com/jaganathanb/dapps-api/pkg/logging"
	"github.com/jaganathanb/dapps-api/pkg/service_errors"
	"gorm.io/gorm"
)

type GstService struct {
	logger       logging.Logger
	cfg          *config.Config
	otpService   *OtpService
	tokenService *TokenService
	database     *gorm.DB
}

func NewGstService(cfg *config.Config) *GstService {
	database := db.GetDb()
	logger := logging.NewLogger(cfg)
	return &GstService{
		cfg:          cfg,
		database:     database,
		logger:       logger,
		otpService:   NewOtpService(cfg),
		tokenService: NewTokenService(cfg),
	}
}

func (s *GstService) CreateGST(req *dto.CreateGSTRequest) error {
	g := models.Gst{Gstin: req.Gstin, TradeName: req.TradeName, RegistrationDate: req.RegistrationDate,
		Locked: false, Address: req.Address, MobileNumber: req.MobileNumber, GSTStatuses: mapGSTStatus(req.GSTStatuses)}

	exists, err := s.existsByGstin(req.Gstin)
	if err != nil {
		return err
	}
	if exists {
		return &service_errors.ServiceError{EndUserMessage: service_errors.GstNotFound}
	}

	tx := s.database.Begin()
	err = tx.Create(&g).Error
	if err != nil {
		tx.Rollback()
		s.logger.Error(logging.Postgres, logging.Rollback, err.Error(), nil)
		return err
	}

	tx.Commit()
	return nil
}

func (s *GstService) existsByGstin(gstin string) (bool, error) {
	var exists bool
	if err := s.database.Model(&models.Gst{}).
		Select("count(*) > 0").
		Where("Gstin = ?", gstin).
		Find(&exists).
		Error; err != nil {
		s.logger.Error(logging.Postgres, logging.Select, err.Error(), nil)
		return false, err
	}
	return exists, nil
}

func mapGSTStatus(statuses []dto.GSTStatus) []models.GstStatus {
	gstatus := make([]models.GstStatus, 0)

	for _, v := range statuses {
		payload := models.GstStatus{
			GstRType:       v.GstRType,
			Status:         v.Status,
			FiledDate:      v.FiledDate,
			PendingReturns: v.PendingReturns,
			TaxPeriod:      v.TaxPeriod,
			Notes:          v.Notes,
		}

		gstatus = append(gstatus, payload)
	}

	return gstatus
}
