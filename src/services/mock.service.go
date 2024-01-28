package services

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/pkg/logging"
)

type MockService struct {
	logger     logging.Logger
	cfg        *config.Config
	httpClient http.Client
}

func NewMockService(cfg *config.Config) *MockService {
	logger := logging.NewLogger(cfg)
	client := http.Client{}

	return &MockService{logger: logger, cfg: cfg, httpClient: client}
}

func (s *MockService) GetMockData(fileName string, prop string) (interface{}, error) {
	fileHandle, err := os.OpenFile(filepath.Join("data/db/mocks", fileName), os.O_RDONLY, os.ModeDevice)

	if err != nil {
		s.logger.Error(logging.Category(logging.IO), logging.SubCategory(logging.OpenFile), err.Error(), nil)
		return nil, err
	}

	defer fileHandle.Close()

	fileBytes, err := io.ReadAll(fileHandle)

	if err != nil {
		s.logger.Error(logging.Category(logging.IO), logging.SubCategory(logging.ReadFile), err.Error(), nil)

		return nil, err
	}

	var mock map[string]interface{}
	json.Unmarshal(fileBytes, &mock)

	return mock[prop], nil
}
