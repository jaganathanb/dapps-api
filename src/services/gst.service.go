package services

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/jaganathanb/dapps-api/api/dto"
	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/constants"
	"github.com/jaganathanb/dapps-api/data/db"
	"github.com/jaganathanb/dapps-api/data/models"
	gst_scrapper "github.com/jaganathanb/dapps-api/pkg/gst-scrapper"
	"github.com/jaganathanb/dapps-api/pkg/logging"
	service_errors "github.com/jaganathanb/dapps-api/pkg/service-errors"
	"github.com/jaganathanb/dapps-api/pkg/utils"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type GstService struct {
	base            *BaseService[models.Gst, dto.CreateGstRequest, dto.UpdateGstReturnStatusRequest, dto.GetGstResponse]
	scrapperService *ScrapperService
	streamerService *StreamerService
	scrapperRunning []string
}

var gstService *GstService
var gstServiceOnce sync.Once

func NewGstService(cfg *config.Config) *GstService {
	gstServiceOnce.Do(func() {
		gstService = &GstService{
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
			scrapperService: NewScrapperService(cfg),
			streamerService: NewStreamerService(cfg),
		}
	})

	return gstService
}

func (s *GstService) CreateGsts(req *dto.CreateGstsRequest) (string, error) {
	exists, err := s.getExistingGstsInSystem(req.Gsts)
	if err != nil {
		return "", err
	}

	gsts := []dto.Gst{}
	for _, v := range req.Gsts {
		if slices.Contains(exists, v.Gstin) {
			s.base.Logger.Warn(logging.Sqlite3, logging.Select, fmt.Sprintf(service_errors.GstExists, v.Gstin), nil)
		} else {
			gsts = append(gsts, v)
		}
	}

	tx := s.base.Database.Begin()

	for _, gst := range gsts {
		err = tx.Create(&models.Gst{
			Gstin:        gst.Gstin,
			MobileNumber: gst.MobileNumber,
			Name:         gst.Name,
			Tradename:    gst.TradeName,
			Email:        gst.Email,
			Type:         gst.Type,
		}).Error
		if err != nil {
			tx.Rollback()
			s.base.Logger.Error(logging.Sqlite3, logging.Rollback, err.Error(), nil)
			return "", err
		}
	}

	tx.Commit()

	gstins := lo.Map(gsts, func(g dto.Gst, i int) string { return g.Gstin })

	go s.scrapGstPortal()

	return fmt.Sprintf("%s gst details already exists. %s gst details entered into the system.", exists, gstins), err
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

	if (req.Status == constants.InvoiceEntry && req.ReturnType == constants.GSTR1) || (req.Status == constants.TaxAmountReceived && req.ReturnType == constants.GSTR3B) {
		go s.scrapGstPortal()
	} else {
		s.scrapperService.streamer.StreamData("Either all GSTs are up-to-date or none of the GSTs are ready to be filed")
	}

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

	err = tx.Model(&models.Gst{}).Where("gstin = ?", req.Gstin).Update("locked", req.Locked).Update("modified_at", time.Now()).Error
	if err != nil {
		tx.Rollback()
		s.base.Logger.Error(logging.Sqlite3, logging.Rollback, err.Error(), nil)
		return false, err
	}

	tx.Commit()

	return true, nil
}

func (s *GstService) DeleteGstById(req *dto.RemoveGstRequest) (bool, error) {
	exists, err := s.isGstExistsInSystem(req.Gstin)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, &service_errors.ServiceError{EndUserMessage: fmt.Sprintf(service_errors.GstNotFound, req.Gstin)}
	}

	tx := s.base.Database.Begin()

	err = tx.Model(&models.GstStatus{}).Where("gstin = ?", req.Gstin).Delete(&models.GstStatus{Gstin: req.Gstin}).Error
	if err != nil {
		tx.Rollback()
		s.base.Logger.Error(logging.Sqlite3, logging.Rollback, err.Error(), nil)
		return false, err
	}

	gst := &models.Gst{}
	err = tx.Model(&models.Gst{}).Where("gstin = ?", req.Gstin).Find(gst).Error
	if err != nil {
		tx.Rollback()
		s.base.Logger.Error(logging.Sqlite3, logging.Rollback, err.Error(), nil)
		return false, err
	}

	err = tx.Delete(gst).Error
	if err != nil {
		tx.Rollback()
		s.base.Logger.Error(logging.Sqlite3, logging.Rollback, err.Error(), nil)
		return false, err
	}

	tx.Commit()

	return true, nil
}

func (s *GstService) GetGstStatistics() (dto.GstFiledCount, error) {
	var totalGstsCount int64
	s.base.Database.Model(&models.Gst{}).Where("locked = ?", false).Count(&totalGstsCount)

	var gstFiledCount dto.GstFiledCount

	err := s.base.Database.Model(&models.GstStatus{}).
		Select(
			"COUNT(CASE WHEN status = 'Filed' AND rtntype = 'GSTR1' THEN 1 END) AS GSTR1Count, " +
				"COUNT(CASE WHEN status = 'Filed' AND rtntype = 'GSTR3B' THEN 1 END) AS GSTR2Count, " +
				"COUNT(CASE WHEN status = 'Filed' AND rtntype = 'GSTR2' THEN 1 END) AS GSTR3BCount, " +
				"COUNT(CASE WHEN status = 'Filed' AND rtntype = 'GSTR9' THEN 1 END) AS GSTR9Count").
		Scan(&gstFiledCount).Error

	gstFiledCount.TotalGsts = totalGstsCount

	return gstFiledCount, err
}

func (s *GstService) RefreshGstReturns() error {
	if s.base.Config.Server.Gst.Username == "" || s.base.Config.Server.Gst.Password == "" {
		return fmt.Errorf("GST Server login details are not correct!")
	}
	go s.scrapGstPortal()

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

func (s *GstService) getExistingGstsInSystem(gsts []dto.Gst) ([]string, error) {
	var exists []string
	if err := s.base.Database.Model(&models.Gst{}).
		Select("Gstin").
		Where("Gstin IN ?", lo.Map(gsts, func(g dto.Gst, i int) string { return g.Gstin })).
		Find(&exists).
		Error; err != nil {
		s.base.Logger.Error(logging.Sqlite3, logging.Select, err.Error(), nil)
		return nil, err
	}

	return exists, nil
}

func (s *GstService) scrapGstPortal() {
	var gsts []models.Gst

	var gstDetail = gst_scrapper.GstDetail{}

	err := s.base.Database.Joins("LEFT JOIN gst_statuses ON gsts.gstin = gst_statuses.gstin").
		Where("gsts.locked = ?", false).
		Where("gst_statuses.gstin IS NULL OR (gst_statuses.status = ? AND gst_statuses.rtntype = ?) OR (gst_statuses.status = ? AND gst_statuses.rtntype = ?) AND gsts.locked = ?", "InvoiceEntry", "GSTR1", "TaxAmountReceived", "GSTR3B", false).
		Preload("GstStatuses").
		Find(&gsts).Error

	if err != nil {
		s.base.Logger.Errorf("Could not query database. %s", err.Error())
	}

	gstins := lo.Map(gsts, func(gst models.Gst, i int) string { return gst.Gstin })

	left, right := lo.Difference(s.scrapperRunning, gstins)

	gstins = slices.Concat(left, right)
	s.scrapperRunning = gstins

	if len(gstins) > 0 {
		s.streamerService.StreamData(fmt.Sprintf("GSTs %s scheduled for return status update", gstins))

		quit := s.scrapperService.ScrapSite(gstins)

		go func() {
			for {
				select {
				case details, ok := <-quit:
					if ok {
						gstDetail = details
						s.updateGstAndReturns(gsts, gstDetail)

						fmt.Printf("Got result for GSTIN %s", gstDetail.Gst.Gstin)
					} else {
						s.base.Logger.Warn(logging.IO, logging.Api, "Done with scrapping!", nil)
						s.scrapperRunning = []string{}
						return
					}
				}
			}
		}()

		fmt.Printf("Total records: %d", len(gsts))

		s.base.Logger.Infof("Job scheduled to update %d GSTs", len(gsts))
	} else {
		s.streamerService.StreamData("Either all GSTs are up-to-date or none of the GSTs are ready to be filed")
	}
}

func (s *GstService) updateGstAndReturns(gsts []models.Gst, gstDetail gst_scrapper.GstDetail) {
	gst, found := lo.Find(gsts, func(gst models.Gst) bool { return gst.Gstin == gstDetail.Gst.Gstin })
	if found {
		tx := s.base.Database.Begin()

		gstDetail.Gst.MobileNumber = gst.MobileNumber
		gstDetail.Gst.Email = gst.Email
		gstDetail.Gst.ModifiedAt = time.Now()

		err := tx.Model(&models.Gst{}).Where("gstin = ?", gst.Gstin).Updates(gstDetail.Gst).Error

		if err != nil {
			tx.Rollback()
			s.base.Logger.Error(logging.Sqlite3, logging.Rollback, err.Error(), nil)
		} else {
			returns := processGstStatuses(gst.Gstin, gstDetail.Returns, len(gst.GstStatuses) == 0)

			err := updateGstReturns(returns, gst, tx)
			if err != nil {
				tx.Rollback()
				s.base.Logger.Error(logging.Sqlite3, logging.Rollback, err.Error(), nil)
			} else {
				tx.Commit()

				s.streamerService.StreamData(fmt.Sprintf("REFRESH_GSTS_TABLE|Gst and its return filing status got update into the system."))
			}
		}
	}
}

func updateGstReturns(returns []models.GstStatus, gst models.Gst, tx *gorm.DB) error {
	var err error
	for _, rtn := range returns {
		if len(gst.GstStatuses) > 0 {
			rtn.ModifiedAt = time.Now()
			err = tx.Model(&models.GstStatus{}).Where("gstin = ? AND rtntype = ?", gst.Gstin, rtn.Rtntype).Updates(&rtn).Error
		} else {
			err = tx.Model(&models.GstStatus{}).Create(&rtn).Error
		}
	}
	return err
}

func processGstStatuses(gstin string, returns []models.GstStatus, isNewEntry bool) []models.GstStatus {
	returnGroups := lo.GroupBy(returns, func(ret models.GstStatus) constants.GstReturnType { return ret.Rtntype })

	newReturns := []models.GstStatus{}
	for rty, retn := range returnGroups {
		rtns := getLatestReturnStatus(rty, retn, isNewEntry)
		if rtns != nil {
			rtns.Gstin = gstin
			newReturns = append(newReturns, *rtns)
		}
	}

	return newReturns
}

func getLatestReturnStatus(gstReturnType constants.GstReturnType, returns []models.GstStatus, isNewEntry bool) *models.GstStatus {
	filed := lo.FilterMap(returns, func(ret models.GstStatus, i int) (models.GstStatus, bool) {
		return models.GstStatus{
				Dof:           ret.Dof,
				RetPrd:        getRetPrdFromTaxp(ret.TaxPrd, ret.FinancialYear, gstReturnType),
				TaxPrd:        ret.TaxPrd,
				FinancialYear: ret.FinancialYear,
				Rtntype:       ret.Rtntype,
				Arn:           ret.Arn,
				Mof:           ret.Mof,
			},
			ret.Status == constants.Filed
	})

	if len(filed) > 0 {
		return getUpdateGstReturnStatus(filed, gstReturnType, isNewEntry)
	} else {
		fmt.Printf("No returns found!")
	}

	return nil
}

func getUpdateGstReturnStatus(filed []models.GstStatus, gstReturnType constants.GstReturnType, isNewEntry bool) *models.GstStatus {
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

	newReturnsStatus := getGstReturn(filed, pendingCount, lastTaxPeriod, lastFiledDate, dueDays, isNewEntry)

	return &newReturnsStatus
}

func getGstReturn(filed []models.GstStatus, pendingCount int, lastTaxPeriod time.Time, lastFiledDate time.Time, dueDays int, isNewEntry bool) models.GstStatus {

	newReturnsStatus := models.GstStatus{TaxPrd: filed[0].TaxPrd, Rtntype: filed[0].Rtntype,
		FinancialYear: filed[0].FinancialYear, Arn: filed[0].Arn, Mof: filed[0].Mof}
	pendings := []string{}

	if pendingCount > 0 {
		for count := range pendingCount {
			if newReturnsStatus.Rtntype == constants.GSTR9 {
				pendings = append(pendings, lastTaxPeriod.AddDate(count+1, 0, 0).Format(constants.TAXPRD))
			} else {
				pendings = append(pendings, lastTaxPeriod.AddDate(0, count+1, 0).Format(constants.TAXPRD))
			}
		}

		newReturnsStatus.Dof = ""
		newReturnsStatus.RetPrd = lastTaxPeriod.AddDate(0, 1, 0).Format(constants.TAXPRD)
		newReturnsStatus.TaxPrd = time.Month(lastTaxPeriod.Month() + 1).String()

		if isNewEntry {
			newReturnsStatus.Status = constants.CallForInvoice
		}
	} else {
		if lastFiledDate.Before(utils.StartOfMonth(time.Now()).AddDate(0, 1, dueDays)) {
			newReturnsStatus.Dof = filed[0].Dof
			newReturnsStatus.RetPrd = filed[0].RetPrd
			newReturnsStatus.TaxPrd = lastTaxPeriod.Month().String()
			newReturnsStatus.Status = constants.Filed
		} else {
			newReturnsStatus.Dof = ""
			newReturnsStatus.RetPrd = lastTaxPeriod.AddDate(0, 1, 0).Format(constants.TAXPRD)
			newReturnsStatus.TaxPrd = lastTaxPeriod.Month().String()
			newReturnsStatus.Status = constants.CallForInvoice
		}
	}

	newReturnsStatus.PendingReturns = pendings

	return newReturnsStatus
}

func getRetPrdFromTaxp(taxp, fy string, returnType constants.GstReturnType) string {
	years := strings.Split(fy, "-")

	if returnType == constants.GSTR9 {
		return fmt.Sprintf("%s%s", fmt.Sprintf("%02d", 2), years[1])
	} else {
		index := utils.Months[taxp]
		if index >= 0 && 2 >= index {
			return fmt.Sprintf("%s%s", fmt.Sprintf("%02d", index), years[1])
		} else {
			return fmt.Sprintf("%s%s", fmt.Sprintf("%02d", index), years[0])
		}
	}
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
