package services

import (
	"net/http"
	"sync"

	"github.com/jaganathanb/dapps-api/common"
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
	streamer   *StreamerService
	scrapper   *gst_scrapper.GstScrapper
}

var scrapperService *ScrapperService
var scrapperServiceOnce sync.Once

func NewScrapperService(cfg *config.Config) *ScrapperService {
	scrapperServiceOnce.Do(func() {
		DB := db.GetDb()
		logger := logging.NewLogger(cfg)
		client := http.Client{}
		streamer := NewStreamerService(cfg)
		scrapper := gst_scrapper.NewGstScrapper(cfg)

		scrapperService = &ScrapperService{logger: logger, cfg: cfg, httpClient: client, DB: DB, streamer: streamer, scrapper: scrapper}
	})

	return scrapperService
}

func (s *ScrapperService) ScrapGstSite(gsts []models.Gst, useCredentialFromSettings bool) (*common.SafeChannel[gst_scrapper.GstDetail], error) {
	return s.scrapper.ScrapGstReturnsDetail(gsts, useCredentialFromSettings)
}
