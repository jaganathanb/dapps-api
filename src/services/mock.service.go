package services

import (
	"net/http"
	"sync"

	"github.com/jaganathanb/dapps-api/config"
	gst_scrapper "github.com/jaganathanb/dapps-api/pkg/gst-scrapper"
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
	quit, gstCh, returnsCh := gst_scrapper.ScrapGstPortal(s.logger)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for {
			select {
			case gst, ok := <-gstCh:
				if ok {
					println(gst)
				} else {
					println("Error !")
				}
			case returns, ok := <-returnsCh:
				if ok {
					println(returns)
				} else {
					println("Error !")
				}
			case <-quit:
				wg.Done()
				return
			}
		}
	}()

	wg.Wait()

	return "", nil
}
