package services

import (
	"net/http"

	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/pkg/logging"
)

type StreamerService struct {
	logger     logging.Logger
	cfg        *config.Config
	httpClient http.Client
}

func NewStreamerService(cfg *config.Config) *StreamerService {
	logger := logging.NewLogger(cfg)
	client := http.Client{}

	return &StreamerService{logger: logger, cfg: cfg, httpClient: client}
}

func (s *StreamerService) StreamData(fileName string, prop string) (interface{}, error) {
	return "", nil
}
