package services

import (
	"fmt"
	"net/http"

	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/data/db"
	"github.com/jaganathanb/dapps-api/data/models"
	gst_scrapper "github.com/jaganathanb/dapps-api/pkg/gst-scrapper"
	"github.com/jaganathanb/dapps-api/pkg/logging"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type ScrapperService struct {
	logger     logging.Logger
	cfg        *config.Config
	httpClient http.Client
	DB         *gorm.DB
}

func NewScrapperService(cfg *config.Config) *ScrapperService {
	DB := db.GetDb()
	logger := logging.NewLogger(cfg)
	client := http.Client{}

	return &ScrapperService{logger: logger, cfg: cfg, httpClient: client, DB: DB}
}

func (s *ScrapperService) ScrapSite() (string, error) {
	var gsts []models.Gst

	var gstDetail = gst_scrapper.GstDetail{}

	err := s.DB.Not("gstin IN (?)", s.DB.Model(&models.GstStatus{}).Select("gstin")).Find(&gsts).Error

	if err != nil {
		return "", err
	}

	quit := gst_scrapper.ScrapGstPortal(lo.Map(gsts, func(gst models.Gst, i int) string { return gst.Gstin }), s.logger, s.cfg)

	go func() {
		for {
			select {
			case details, ok := <-quit:
				if ok {
					gstDetail = details
					updateGstAndReturns(gsts, gstDetail, s)

					fmt.Printf("Got result for GSTIN %s", gstDetail.Gst.Gstin)
				} else {
					println("Error !")
					return
				}
			}
		}
	}()

	fmt.Printf("Totla records: %d", len(gsts))

	return fmt.Sprintf("Job scheduled to update %d GSTs", len(gsts)), nil
}

func updateGstAndReturns(gsts []models.Gst, gstDetail gst_scrapper.GstDetail, s *ScrapperService) {
	gst, found := lo.Find(gsts, func(gst models.Gst) bool { return gst.Gstin == gstDetail.Gst.Gstin })
	if found {
		tx := s.DB.Begin()

		gstDetail.Gst.MobileNumber = gst.MobileNumber
		gstDetail.Gst.Email = gst.Email

		err := tx.Model(&gst).Updates(gstDetail.Gst).Error

		if err != nil {
			tx.Rollback()
			s.logger.Error(logging.Sqlite3, logging.Rollback, err.Error(), nil)
		} else {
			err := tx.Model(&models.GstStatus{}).Create(processGstStatuses(gst.Gstin, gstDetail.Returns)).Error
			if err != nil {
				tx.Rollback()
				s.logger.Error(logging.Sqlite3, logging.Rollback, err.Error(), nil)
			} else {
				tx.Commit()
			}
		}
	}
}
