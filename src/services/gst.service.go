package services

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/jaganathanb/dapps-api/api/dto"
	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/constants"
	"github.com/jaganathanb/dapps-api/data/db"
	"github.com/jaganathanb/dapps-api/data/models"
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

	gstins := []string{}
	for _, v := range req.Gstins {
		if slices.Contains(exists, v) {
			s.base.Logger.Warn(logging.Sqlite3, logging.Select, fmt.Sprintf(service_errors.GstExists, v), nil)
		} else {
			gstins = append(gstins, v)
		}
	}

	tx := s.base.Database.Begin()

	for _, gstin := range gstins {
		err = tx.Create(&models.Gst{Gstin: gstin}).Error
		if err != nil {
			tx.Rollback()
			s.base.Logger.Error(logging.Sqlite3, logging.Rollback, err.Error(), nil)
			return "", err
		}
	}

	tx.Commit()

	return fmt.Sprintf("%s gst details already exists. %s gst details entered into the system.", exists, gstins), err
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
		rtns := getLatestReturnStatus(rty, retn, gstin)
		if rtns != nil {
			newReturns = append(newReturns, *rtns)
		}
	}

	return newReturns
}

func getLatestReturnStatus(gstReturnType constants.GstReturnType, returns []models.GstStatus, gstin string) *models.GstStatus {
	pendings := []string{}

	filed := lo.FilterMap(returns, func(ret models.GstStatus, i int) (models.GstStatus, bool) {
		return models.GstStatus{
				Dof:           ret.Dof,
				RetPrd:        getRetPrdFromTaxp(ret.TaxPrd, ret.FinancialYear),
				TaxPrd:        ret.TaxPrd,
				FinancialYear: ret.FinancialYear,
				Arn:           ret.Arn,
				Mof:           ret.Mof,
			},
			ret.Status == constants.Filed
	})

	if len(filed) > 0 {
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

		return &newReturnsStatus
	} else {
		fmt.Printf("No returns found for GSTIN %s", gstin)
	}

	return nil
}

func getRetPrdFromTaxp(taxp, fy string) string {
	currYearMonths := constants.GetMonthNames()[:3]
	years := strings.Split(fy, "-")

	index := lo.IndexOf(currYearMonths, taxp)
	if index > -1 {
		return fmt.Sprintf("%s%s", fmt.Sprint(index), years[0])
	} else {
		return fmt.Sprintf("%s%s", fmt.Sprint(index), years[1])
	}
}

func (s *GstService) GetByFilter(ctx context.Context, req *dto.PaginationInputWithFilter) (*dto.PagedList[dto.GetGstResponse], error) {
	return s.base.GetByFilter(ctx, req)
}

func (s *GstService) UpdateGstStatus(req *dto.UpdateGstReturnStatusRequest) (bool, error) {
	exists, err := s.isGstExistsInSystem(req.Gstin)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, &service_errors.ServiceError{EndUserMessage: fmt.Sprintf(service_errors.GstNotFound, req.Gstin)}
	}

	tx := s.base.Database.Begin()

	err = tx.Model(&models.GstStatus{}).Where("gstin = ? AND rtntype = ?", req.Gstin, req.ReturnType).Update("status", req.Status).Error
	if err != nil {
		tx.Rollback()
		s.base.Logger.Error(logging.Sqlite3, logging.Rollback, err.Error(), nil)
		return false, err
	}

	tx.Commit()
	return true, nil
}

func (s *GstService) LockGstById(req *dto.UpdateGstLockStatusRequest) (bool, error) {
	exists, err := s.isGstExistsInSystem(req.Gstin)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, &service_errors.ServiceError{EndUserMessage: fmt.Sprintf(service_errors.GstNotFound, req.Gstin)}
	}

	tx := s.base.Database.Begin()

	err = tx.Model(&models.Gst{}).Where("gstin = ?", req.Gstin).Update("locked", req.Locked).Error
	if err != nil {
		tx.Rollback()
		s.base.Logger.Error(logging.Sqlite3, logging.Rollback, err.Error(), nil)
		return false, err
	}

	tx.Commit()
	return true, nil
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
