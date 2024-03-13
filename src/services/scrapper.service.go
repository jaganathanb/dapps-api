package services

import (
	"fmt"
	"net/http"

	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/data/db"
	"github.com/jaganathanb/dapps-api/data/models"
	gst_scrapper "github.com/jaganathanb/dapps-api/pkg/gst-scrapper"
	"github.com/jaganathanb/dapps-api/pkg/logging"
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

func (s *ScrapperService) ScrapSite() error {
	var gsts []models.Gst

	err := s.DB.Not("gstin IN (?)", s.DB.Model(&models.GstStatus{}).Select("gstin")).Find(&gsts).Error

	if err != nil {
		return err
	}

	for _, gst := range gsts {
		gst_scrapper.ScrapGstPortal(gst.Gstin, s.logger, s.cfg)
	}

	fmt.Printf("Totla records: %d", len(gsts))

	return nil
}
