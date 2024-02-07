package services

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/jaganathanb/dapps-api/api/dto"
	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/constants"
	"github.com/jaganathanb/dapps-api/data/db"
	"github.com/jaganathanb/dapps-api/data/models"
	httpwrapper "github.com/jaganathanb/dapps-api/pkg/http-wrapper"
	"github.com/jaganathanb/dapps-api/pkg/logging"
	service_errors "github.com/jaganathanb/dapps-api/pkg/service-errors"
	"github.com/samber/lo"
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
				{string: "PAddress"},
			},
			Config: cfg,
		},
	}
}

func (s *GstService) CreateGsts(req *dto.CreateGstsRequest) (string, error) {
	exists, err := s.getExistingGstsInSystem(req.Gstins)
	if err != nil {
		return "", err
	}

	var reqs []http.Request
	for _, v := range req.Gstins {
		if slices.Contains[[]string](exists, v) {
			s.base.Logger.Warn(logging.Sqlite3, logging.Select, fmt.Sprintf(service_errors.GstExists, v), nil)
		} else {
			var url string
			if s.base.Config.Server.RunMode == "release" {
				url = fmt.Sprintf("https://taxpayer.irisgst.com/api/search?gstin=%s", v)
			} else {
				url = fmt.Sprintf("http://localhost:%s/api/v%d/mocks/gsts/%s", s.base.Config.Server.ExternalPort, constants.Version, v)
			}
			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				s.base.Logger.Error(logging.Category(logging.ExternalService), logging.Api, err.Error(), nil)
			} else {
				reqs = append(reqs, *req)
			}
		}
	}

	if len(reqs) == 0 {
		s.base.Logger.Warn(logging.Sqlite3, logging.Insert, fmt.Sprintf(service_errors.GstsExists, req.Gstins), nil)

		return "No gst entered into tothe system", nil
	}

	res, _ := httpwrapper.AsyncHTTP[models.Gst](reqs)

	if len(res) > 0 && lo.SomeBy(res, func(r dto.HttpResonseWrapper[models.Gst]) bool { return r.Data != nil }) {
		tx := s.base.Database.Begin()

		for _, v := range res {
			if v.Error != nil {
				s.base.Logger.Error(logging.Sqlite3, logging.Rollback, v.Error.Error(), nil)
			} else {
				err = tx.Create(&v.Data.Result).Error
				if err != nil {
					tx.Rollback()
					s.base.Logger.Error(logging.Sqlite3, logging.Rollback, err.Error(), nil)
					return "", err
				}
			}
		}

		tx.Commit()
	}

	return fmt.Sprintf("%s gst details already exists. %d gst details entered into the system.", exists, len(res)), nil
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
