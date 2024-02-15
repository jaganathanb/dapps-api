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
				{string: "Pradr"},
				{string: "Adadr"},
				{string: "Addr"},
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

	reqs := map[string][]dto.HttpRequestConfig{}
	for _, v := range req.Gstins {
		if slices.Contains[[]string](exists, v) {
			s.base.Logger.Warn(logging.Sqlite3, logging.Select, fmt.Sprintf(service_errors.GstExists, v), nil)
		} else {
			var gstUrl, returnsUrl string
			if s.base.Config.Server.RunMode == "release" {
				gstUrl = fmt.Sprintf("https://taxpayer.irisgst.com/api/search?gstin=%s", v)
				returnsUrl = fmt.Sprintf("https://taxpayer.irisgst.com/api/returnstatus?gstin=%s", v)
			} else {
				gstUrl = fmt.Sprintf("http://localhost:%s/api/v%d/mocks/gsts/%s", s.base.Config.Server.ExternalPort, constants.Version, v)
				returnsUrl = fmt.Sprintf("http://localhost:%s/api/v%d/mocks/returns/%s", s.base.Config.Server.ExternalPort, constants.Version, v)
			}

			rqs := []dto.HttpRequestConfig{{Method: http.MethodGet, RequestID: v, URL: gstUrl, ResponseType: &dto.HttpResponseResult[models.Gst]{}}, {Method: http.MethodGet, RequestID: v, URL: returnsUrl, ResponseType: &dto.HttpResponseResult[[]models.GstStatus]{}}}
			reqs[v] = rqs
		}
	}

	client := httpwrapper.NewHTTPClient()
	for _, req := range reqs {
		responses := client.MakeRequests(req)

		// Process responses
		s.newMethod(responses)
	}

	return fmt.Sprintf("%s gst details already exists. %d gst details entered into the system.", exists, 2), nil
}

func (*GstService) newMethod(responses []dto.HttpResponseWrapper) {
	grouped := lo.GroupBy(responses, func(res dto.HttpResponseWrapper) string { return res.RequestID })

	for _, resps := range grouped {
		respsWithoutErr := lo.Filter(resps, func(res dto.HttpResponseWrapper, i int) bool { return res.Err == nil })

		if len(respsWithoutErr) == 2 {
			var gst models.Gst
			var returns []models.GstStatus
			for _, res := range resps {
				switch res.ResponseType.(type) {
				case *models.Gst:
					if rType, ok := res.Body.(*models.Gst); ok {
						gst = *rType
					}
				case *[]models.GstStatus:
					if rType, ok := res.Body.(*[]models.GstStatus); ok {
						returns = *rType
					}
				}
			}
			fmt.Println(gst)
			fmt.Println(returns)
		}

	}
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
			Rtntype:        v.GstRType,
			Status:         v.Status,
			Dof:            v.FiledDate,
			PendingReturns: v.PendingReturns,
			RetPrd:         v.TaxPeriod,
			Notes:          v.Notes,
		}

		gstatus = append(gstatus, payload)
	}

	return gstatus
}
