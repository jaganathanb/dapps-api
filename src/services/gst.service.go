package services

import (
	"context"

	"github.com/jaganathanb/dapps-api/api/dto"
	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/data/db"
	"github.com/jaganathanb/dapps-api/data/models"
	"github.com/jaganathanb/dapps-api/pkg/logging"
	"github.com/jaganathanb/dapps-api/pkg/service_errors"
	"gorm.io/gorm"
)

type GstService struct {
	base *BaseService[models.Gst, dto.CreateGSTRequest, dto.UpdateGSTReturnStatusRequest, dto.GetGstResponse]
}

func NewGstService(cfg *config.Config) *GstService {

	return &GstService{
		base: &BaseService[models.Gst, dto.CreateGSTRequest, dto.UpdateGSTReturnStatusRequest, dto.GetGstResponse]{
			Database: db.GetDb(),
			Logger:   logging.NewLogger(cfg),
			Preloads: []preload{
				{string: "GstStatuses"},
			},
		},
	}
}

func (s *GstService) CreateGst(req *dto.CreateGSTRequest) error {
	g := models.Gst{Gstin: req.Gstin, TradeName: req.TradeName, RegistrationDate: req.RegistrationDate,
		Locked: false, Address: req.Address, MobileNumber: req.MobileNumber, GstStatuses: mapGSTStatus(req.GSTStatuses)}

	exists, err := s.existsByGstin(req.Gstin)
	if err != nil {
		return err
	}
	if exists {
		return &service_errors.ServiceError{EndUserMessage: service_errors.GstExists}
	}

	tx := s.base.Database.Begin()
	err = tx.Create(&g).Error
	if err != nil {
		tx.Rollback()
		s.base.Logger.Error(logging.Sqlite3, logging.Rollback, err.Error(), nil)
		return err
	}

	tx.Commit()
	return nil
}

func (s *GstService) GetByFilter(ctx context.Context, req *dto.PaginationInputWithFilter) (*dto.PagedList[dto.GetGstResponse], error) {
	return s.base.GetByFilter(ctx, req)
}

func (s *GstService) UpdateGstStatuses(req *dto.UpdateGSTReturnStatusRequest) error {
	exists, err := s.existsByGstin(req.Gstin)
	if err != nil {
		return err
	}
	if !exists {
		return &service_errors.ServiceError{EndUserMessage: service_errors.GstNotFound}
	}

	tx := s.base.Database.Begin()

	statuses := mapGSTStatus(req.GSTStatuses)

	for _, v := range statuses {
		err = tx.Where("gstin = ?", req.Gstin).Updates(&v).Error
		if err != nil {
			tx.Rollback()
			s.base.Logger.Error(logging.Sqlite3, logging.Rollback, err.Error(), nil)
			return err
		}
	}

	tx.Commit()
	return nil
}

func (s *GstService) existsByGstin(gstin string) (bool, error) {
	var exists bool
	if err := s.base.Database.Model(&models.Gst{}).
		Select("count(*) > 0").
		Where("Gstin = ?", gstin).
		Preload("GstStatuses", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("GstStatus")
		}).
		Find(&exists).
		Error; err != nil {
		s.base.Logger.Error(logging.Sqlite3, logging.Select, err.Error(), nil)
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
