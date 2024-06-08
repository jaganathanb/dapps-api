package services

import (
	"context"
	"fmt"
	"math"
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
	base                 *BaseService[models.Gst, dto.CreateGstRequest, dto.UpdateGstReturnStatusRequest, dto.GetGstResponse]
	scrapperService      *ScrapperService
	streamerService      *StreamerService
	notificationsService *NotificationsService
	scrapperRunning      []string
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
			scrapperService:      NewScrapperService(cfg),
			streamerService:      NewStreamerService(cfg),
			notificationsService: NewNotificationsService(cfg),
		}
	})

	return gstService
}

func (s *GstService) CreateGsts(req *dto.CreateGstsRequest) (string, error) {
	exists, err := s.getExistingGstsInSystem(req.Gsts)
	if err != nil {
		return "", err
	}

	tx := s.base.Database.Begin()

	gsts := []dto.Gst{}
	for _, v := range req.Gsts {
		payload := &models.Gst{
			Sno:          v.Sno,
			Fno:          v.Fno,
			Gstin:        v.Gstin,
			MobileNumber: v.MobileNumber,
			Name:         v.Name,
			Tradename:    v.TradeName,
			Email:        v.Email,
			Type:         v.Type,
		}

		if slices.Contains(exists, v.Gstin) {
			err = tx.Updates(payload).Error
			if err != nil {
				tx.Rollback()
				s.base.Logger.Error(logging.Sqlite3, logging.Rollback, err.Error(), nil)
				return "", err
			}
			s.base.Logger.Warn(logging.Sqlite3, logging.Update, fmt.Sprintf(service_errors.GstExists, v.Gstin), nil)
		} else {
			err = tx.Create(payload).Error
			if err != nil {
				tx.Rollback()
				s.base.Logger.Error(logging.Sqlite3, logging.Rollback, err.Error(), nil)
				return "", err
			}
		}
	}

	tx.Commit()

	gstins := lo.Map(gsts, func(g dto.Gst, i int) string { return g.Gstin })

	s.scrapGstPortal(req.CreatedBy)

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

	if (req.ReturnType == constants.GSTR1 && req.Status == constants.InvoiceEntry) || (req.ReturnType == constants.GSTR3B && req.Status == constants.TaxAmountReceived) {
		go s.scrapGstPortal(req.ModifiedBy)
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
	var gstins []string
	s.base.Database.Model(&models.Gst{}).Where("locked = ? AND gsts.status = ?", false, "Active").Select("gstin").Find(&gstins)

	var gstFiledCount dto.GstFiledCount
	var statuses []models.GstStatus

	err := s.base.Database.Model(&models.GstStatus{}).Where("gstin in ?", gstins).Select("pending_returns", "rtntype", "ret_prd", "status", "gstin").Find(&statuses).Error

	if err != nil {
		return gstFiledCount, err
	}

	group := lo.GroupBy(statuses, func(st models.GstStatus) constants.GstReturnType { return st.Rtntype })

	currTxp := time.Now().Format(constants.TAXPRD)
	gstFiledCount.GSTR1Count = getPendingReturnStatus(group[constants.GSTR1], currTxp, constants.GSTR1)
	gstFiledCount.GSTR3BCount = getPendingReturnStatus(group[constants.GSTR3B], currTxp, constants.GSTR3B)
	gstFiledCount.GSTR2Count = getPendingReturnStatus(group[constants.GSTR2], currTxp, constants.GSTR2)
	gstFiledCount.GSTR9Count = getPendingReturnStatus(group[constants.GSTR9], currTxp, constants.GSTR9)

	gstFiledCount.TotalGsts = int64(len(gstins))

	return gstFiledCount, err
}

func getPendingReturnStatus(statuses []models.GstStatus, currTxp string, retType constants.GstReturnType) int64 {
	return int64(len(lo.Filter(statuses, func(st models.GstStatus, i int) bool {
		return len(lo.Filter(st.PendingReturns, func(ss string, j int) bool { return ss != currTxp && st.Rtntype == retType })) > 0
	})))
}

func (s *GstService) RefreshGstReturns(userId int) error {
	go s.scrapGstPortal(userId)

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

func (s *GstService) scrapGstPortal(userId int) {
	var gsts []models.Gst

	var gstDetail = gst_scrapper.GstDetail{}

	if s.base.Config.Server.Gst.BaseUrl == "" || s.base.Config.Server.Gst.Username == "" || s.base.Config.Server.Gst.Password == "" {
		s.streamerService.StreamData(StreamMessage{Message: "GST settings are not available. Please update it from Settings page.", MessageType: constants.ERROR})

		return
	}

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

	gstins = lo.Uniq(slices.Concat(left, right))
	s.scrapperRunning = gstins

	count := len(gstins)

	if count > 0 {
		errMsg := fmt.Sprintf("%d GSTs scheduled for return status update", count)

		if len(gstins) > 5 {
			errMsg = fmt.Sprintf("GSTIN %s and %d+ more has been scheduled for GST Return status", gstins[0], int(math.Floor(float64(count/5)*5)))
		} else if count == 1 {
			errMsg = fmt.Sprintf("GSTIN %s has been for GST Return status", gstins[0])
		} else {
			errMsg = fmt.Sprintf("GSTIN %s and %d more has been GST Return status", gstins[0], count-1)
		}

		s.streamerService.StreamData(StreamMessage{Message: errMsg, UserId: userId})

		quit, err := s.scrapperService.ScrapSite(gstins)
		if err == nil {
			go func() {
				success := []string{}
				failed := 0
				for {
					select {
					case details, ok := <-quit.C:
						if ok {
							if details.ErrorMessage == "" {
								success = append(success, details.Gst.Gstin)
								gstDetail = details
								s.updateGstAndReturns(gsts, gstDetail)

								s.base.Logger.Infof("Got result for GSTIN %s", gstDetail.Gst.Gstin)
							} else {
								failed += 1
								s.base.Logger.Errorf("Failed to fetch data for a GSTIN - %s", details.ErrorMessage)

								messages := strings.Split(details.ErrorMessage, "|")

								if len(messages) > 1 && messages[0] == "NOTIFICATION" {
									errMsg = messages[1]
									s.streamerService.StreamData(StreamMessage{Code: "NOTIFICATION", UserId: userId, MessageType: constants.ERROR, Message: errMsg})
								}
							}
						} else {
							s.base.Logger.Infof("Done with scrapping for %s GSTINs", gstins)
							s.scrapperRunning = []string{}

							count := len(gstins)
							errMsg := ""

							if len(success) == count {
								if len(gstins) > 5 {
									errMsg = fmt.Sprintf("GST Return status for GSTIN %s and %d+ more has been updated into the system", gstins[0], int(math.Floor(float64(count/5)*5)))
								} else if count == 1 {
									errMsg = fmt.Sprintf("GST Return status for GSTIN %s has been updated into the system", gstins[0])
								} else {
									errMsg = fmt.Sprintf("GST Return status for GSTIN %s and %d more has been updated into the system", gstins[0], count-1)
								}

								s.streamerService.StreamData(StreamMessage{Code: "NOTIFICATION", UserId: userId, MessageType: constants.SUCCESS, Message: errMsg})
							} else if failed == count {
								errMsg = fmt.Sprintf("Something went wrong!. Could not process any of the GSTINs submitted. Please check logs for more details.")

								s.streamerService.StreamData(StreamMessage{Code: "NOTIFICATION", UserId: userId, MessageType: constants.ERROR, Message: errMsg})
							} else {
								if len(success) != 0 {
									errMsg = fmt.Sprintf("Something went wrong!. Though system could able to process %d GSTINs successfully and %d GSTINs failed to process", len(success), count-len(success))
									s.streamerService.StreamData(StreamMessage{Code: "REFRESH_GSTS_TABLE"})
								} else {
									errMsg = fmt.Sprintf("Something went wrong!. %d GSTINs failed to process", count)
								}

								s.streamerService.StreamData(StreamMessage{Code: "NOTIFICATION", UserId: userId, MessageType: constants.ERROR, Message: errMsg})
							}

							return
						}
					}
				}
			}()

			fmt.Printf("Total records: %d", len(gsts))

			s.base.Logger.Infof("Job scheduled to update %d GSTs", len(gsts))
		} else {
			s.streamerService.StreamData(StreamMessage{Message: "Something went wrong!. Could not process Gst Returns.", UserId: userId, MessageType: constants.ERROR, Code: "NOTIFICATION"})
			s.scrapperRunning = []string{}
		}

	} else {
		s.streamerService.StreamData(StreamMessage{Message: "Either all GSTs are up-to-date or none of the GSTs are ready to be filed"})
		s.scrapperRunning = []string{}
	}
}

func (s *GstService) updateGstAndReturns(gsts []models.Gst, gstDetail gst_scrapper.GstDetail) {
	gst, found := lo.Find(gsts, func(gst models.Gst) bool { return gst.Gstin == gstDetail.Gst.Gstin })
	if found {
		tx := s.base.Database.Begin()

		gstDetail.Gst.MobileNumber = gst.MobileNumber
		gstDetail.Gst.Email = gst.Email
		gstDetail.Gst.Locked = gstDetail.Gst.Status != "Active"
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
			setReturnStatus(newReturnsStatus)
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
			setReturnStatus(newReturnsStatus)
		}
	}

	newReturnsStatus.PendingReturns = pendings

	return newReturnsStatus
}

func setReturnStatus(newReturnsStatus models.GstStatus) {
	switch newReturnsStatus.Rtntype {
	case constants.GSTR1:
		newReturnsStatus.Status = constants.CallForInvoice
		break
	case constants.GSTR3B:
		newReturnsStatus.Status = constants.TaxPayable
		break
	default:
		newReturnsStatus.Status = constants.CallForInvoice
	}
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
