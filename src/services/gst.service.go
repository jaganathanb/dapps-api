package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/jaganathanb/dapps-api/api/dto"
	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/constants"
	"github.com/jaganathanb/dapps-api/data/db"
	"github.com/jaganathanb/dapps-api/data/models"
	httpwrapper "github.com/jaganathanb/dapps-api/pkg/http-wrapper"
	"github.com/jaganathanb/dapps-api/pkg/logging"
	service_errors "github.com/jaganathanb/dapps-api/pkg/service-errors"
	"github.com/jaganathanb/dapps-api/pkg/utils"
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
				gstUrl = fmt.Sprintf("%ssearch?gstin=%s", s.base.Config.Server.GstBaseUrl, v)
				returnsUrl = fmt.Sprintf("%sreturnstatus?gstin=%s", s.base.Config.Server.GstBaseUrl, v)
			} else {
				gstUrl = fmt.Sprintf("%smocks/gsts/%s", s.base.Config.Server.GstBaseUrl, v)
				returnsUrl = fmt.Sprintf("%smocks/returns/%s", s.base.Config.Server.GstBaseUrl, v)
			}

			rqs := []dto.HttpRequestConfig{{Method: http.MethodGet, RequestID: v, URL: gstUrl, ResponseType: &dto.HttpResponseResult[models.Gst]{}}, {Method: http.MethodGet, RequestID: v, URL: returnsUrl, ResponseType: &dto.HttpResponseResult[[]models.GstStatus]{}}}
			reqs[v] = rqs
		}
	}

	client := httpwrapper.NewHTTPClient(*s.base.Config)

	gsts := []models.Gst{}
	errResps := []dto.HttpResponseWrapper{}
	for _, req := range reqs {
		responses := client.MakeRequests(req)

		// Process responses
		gs, es := processResponses(responses)

		gsts = append(gsts, gs...)
		errResps = append(errResps, es...)

		tx := s.base.Database.Begin()

		for _, gst := range gsts {
			err = tx.Create(&gst).Error
			if err != nil {
				tx.Rollback()
				s.base.Logger.Error(logging.Sqlite3, logging.Rollback, err.Error(), nil)
				return "", err
			}
		}

		tx.Commit()
	}

	ids := lo.Map(gsts, func(gst models.Gst, i int) string { return gst.Gstin })
	errIds := lo.Uniq(lo.Map(errResps, func(err dto.HttpResponseWrapper, i int) string { return err.RequestID }))

	if len(errIds) == len(req.Gstins) {
		err = errors.New(fmt.Sprintf("Something went wrong with %s gstins", errIds))
	}

	return fmt.Sprintf("%s gst details already exists. %s gst details entered into the system. %s gst details got errored out", exists, ids, errIds), err
}

func processResponses(responses []dto.HttpResponseWrapper) ([]models.Gst, []dto.HttpResponseWrapper) {
	grouped := lo.GroupBy(responses, func(res dto.HttpResponseWrapper) string { return res.RequestID })
	gsts := []models.Gst{}
	errResps := []dto.HttpResponseWrapper{}

	for _, resps := range grouped {
		respsWithoutErr := lo.Filter(resps, func(res dto.HttpResponseWrapper, i int) bool { return res.Err == nil })
		errResps = append(errResps, lo.Filter(resps, func(res dto.HttpResponseWrapper, i int) bool { return res.Err != nil })...)

		if len(respsWithoutErr) == 2 {
			var gst models.Gst
			var returns []models.GstStatus
			for _, res := range resps {
				switch res.ResponseType.(type) {
				case *dto.HttpResponseResult[models.Gst]:
					if rType, ok := res.Body.(*dto.HttpResponseResult[models.Gst]); ok {
						gst = *&rType.Result
					}
				case *dto.HttpResponseResult[[]models.GstStatus]:
					if rType, ok := res.Body.(*dto.HttpResponseResult[[]models.GstStatus]); ok {
						returns = *&rType.Result
					}
				}
			}

			if gst.Gstin != "" {
				gst.GstStatuses = processGstStatuses(gst.Gstin, returns)
				gsts = append(gsts, gst)
			}
		}
	}

	return gsts, errResps
}

func processGstStatuses(gstin string, returns []models.GstStatus) []models.GstStatus {
	returnGroups := lo.GroupBy(returns, func(ret models.GstStatus) constants.GstReturnType { return ret.Rtntype })

	newReturns := []models.GstStatus{}
	for rty, retn := range returnGroups {
		newReturns = append(newReturns, getLatestReturnStatus(rty, retn, gstin))
	}

	return newReturns
}

func getLatestReturnStatus(gstReturnType constants.GstReturnType, returns []models.GstStatus, gstin string) models.GstStatus {
	pendings := []string{}

	filed := lo.FilterMap(returns, func(ret models.GstStatus, i int) (models.GstStatus, bool) {
		return models.GstStatus{Dof: ret.Dof, RetPrd: ret.RetPrd, Arn: ret.Arn, Mof: ret.Mof}, ret.Status == constants.Filed
	})

	slices.SortFunc(filed,
		func(a, b models.GstStatus) int {
			dof1, err1 := time.Parse(constants.DOF, a.Dof)
			dof2, err2 := time.Parse(constants.DOF, b.Dof)

			if err1 == nil && err2 == nil {
				if dof1.After(dof2) {
					return -1
				} else {
					return 1
				}
			}

			return 1
		})

	lastFiledDate, _ := time.Parse(constants.DOF, filed[0].Dof)
	lastTaxPeriod, _ := time.Parse(constants.TAXPRD, filed[0].RetPrd)

	years, months, _, _, _, _ := utils.Diff(time.Now(), lastTaxPeriod)

	var pendingCount int
	var dueDays int

	if gstReturnType == constants.GSTR9 {
		pendingCount = years
		dueDays = 21
	} else {
		pendingCount = months
		dueDays = 12
	}

	newReturnsStatus := models.GstStatus{Gstin: gstin, Rtntype: gstReturnType, Arn: filed[0].Arn, Mof: filed[0].Mof}

	if pendingCount > 0 {
		for count := range pendingCount {
			pendings = append(pendings, lastTaxPeriod.AddDate(0, count+1, 0).Format(constants.TAXPRD))
		}

		newReturnsStatus.Dof = ""
		newReturnsStatus.RetPrd = lastTaxPeriod.AddDate(0, 1, 0).Format(constants.TAXPRD)
		newReturnsStatus.Status = constants.InvoiceCall
	} else {
		if lastFiledDate.Before(utils.StartOfMonth(time.Now()).AddDate(0, 1, dueDays)) {
			newReturnsStatus.Dof = filed[0].Dof
			newReturnsStatus.RetPrd = filed[0].RetPrd
			newReturnsStatus.Status = constants.Filed
		} else {
			newReturnsStatus.Dof = ""
			newReturnsStatus.RetPrd = lastTaxPeriod.AddDate(0, 1, 0).Format(constants.TAXPRD)
			newReturnsStatus.Status = constants.InvoiceCall
		}
	}

	newReturnsStatus.PendingReturns = pendings

	return newReturnsStatus
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
			Rtntype:        v.ReturnType,
			Status:         v.Status,
			Dof:            v.LastFiledDate,
			PendingReturns: v.PendingReturns,
			RetPrd:         v.ReturnPeriod,
			Notes:          v.Notes,
		}

		gstatus = append(gstatus, payload)
	}

	return gstatus
}
