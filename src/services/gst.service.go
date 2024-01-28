package services

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/jaganathanb/dapps-api/api/dto"
	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/data/db"
	"github.com/jaganathanb/dapps-api/data/models"
	httpwrapper "github.com/jaganathanb/dapps-api/pkg/http-wrapper"
	"github.com/jaganathanb/dapps-api/pkg/logging"
	service_errors "github.com/jaganathanb/dapps-api/pkg/service-errors"
	"gorm.io/gorm"
)

type GstService struct {
	base *BaseService[models.Gst, dto.CreateGstRequest, dto.UpdateGstReturnStatusRequest, dto.GetGstResponse]
}

func NewGstService(cfg *config.Config) *GstService {

	return &GstService{
		base: &BaseService[models.Gst, dto.CreateGstRequest, dto.UpdateGstReturnStatusRequest, dto.GetGstResponse]{
			Database: db.GetDb(),
			Logger:   logging.NewLogger(cfg),
			Preloads: []preload{
				{string: "GstStatuses"},
			},
		},
	}
}

func (s *GstService) CreateGsts(req *dto.CreateGstsRequest) error {
	exists, err := s.getExistingGstsInSystem(req.Gstins)
	if err != nil {
		return err
	}

	var reqs []http.Request
	for _, v := range req.Gstins {
		if slices.Contains[[]string](exists, v) {
			s.base.Logger.Warn(logging.Sqlite3, logging.Select, fmt.Sprintf(service_errors.GstExists, v), nil)
		} else {
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://taxpayer.irisgst.com/api/search?gstin=%s", v), nil)
			if err != nil {
				s.base.Logger.Error(logging.Category(logging.ExternalService), logging.Api, err.Error(), nil)
			} else {
				reqs = append(reqs, *req)
			}
		}
	}

	res, _ := httpwrapper.AsyncHTTP[dto.GetGstResponse](reqs)

	tx := s.base.Database.Begin()
	for _, v := range res {
		gst := models.Gst{
			Gstin:            v.Gstin,
			TradeName:        v.TradeName,
			RegistrationDate: v.RegistrationDate,
			Locked:           v.Locked,
			Address:          v.Address,
			MobileNumber:     v.MobileNumber,
			GstStatuses:      mapGSTStatus(v.GstStatuses),
		}

		err = tx.Create(&gst).Error
		if err != nil {
			tx.Rollback()
			s.base.Logger.Error(logging.Sqlite3, logging.Rollback, err.Error(), nil)
			return err
		}
	}

	tx.Commit()
	return nil
}

func (s *GstService) GetByFilter(ctx context.Context, req *dto.PaginationInputWithFilter) (*dto.PagedList[dto.GetGstResponse], error) {
	return s.base.GetByFilter(ctx, req)
}

func (s *GstService) UpdateGstStatuses(req *dto.UpdateGstReturnStatusRequest) error {
	exists, err := s.isGstExistsInSystem(req.Gstin)
	if err != nil {
		return err
	}
	if !exists {
		return &service_errors.ServiceError{EndUserMessage: fmt.Sprintf(service_errors.GstNotFound, req.Gstin)}
	}

	tx := s.base.Database.Begin()

	statuses := mapGSTStatus(req.GstStatuses)

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

func (s *GstService) isGstExistsInSystem(gstin string) (bool, error) {
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

func (s *GstService) getExistingGstsInSystem(gstins []string) ([]string, error) {
	var exists []string
	if err := s.base.Database.Model(&models.Gst{}).
		Select("Gstin").
		Where("Gstin IN ?", gstins).
		Find(&exists).
		Error; err != nil {
		s.base.Logger.Error(logging.Sqlite3, logging.Select, err.Error(), nil)
		return nil, err
	}

	return exists, nil
}

func mapGSTStatus(statuses []dto.GstStatus) []models.GstStatus {
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
