package httpwrapper

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/pkg/logging"
)

var log = logging.NewLogger(config.GetConfig())
var client = http.Client{}

func makeCall[R any](req *http.Request, ch chan<- R, wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := client.Do(req)
	if err != nil {
		log.Error(logging.Category(logging.ExternalService), logging.SubCategory(logging.RequestResponse), err.Error(), nil)
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(logging.Category(logging.ExternalService), logging.SubCategory(logging.RequestResponse), err.Error(), nil)
	}

	var res R
	json.Unmarshal(b, &res)

	ch <- R(res)
}

func AsyncHTTP[R any](reqs []http.Request) ([]R, error) {
	ch := make(chan R)
	var responses []R
	var wg sync.WaitGroup

	for _, req := range reqs {
		wg.Add(1)
		go makeCall(&req, ch, &wg)
	}

	// close the channel in the background
	go func() {
		wg.Wait()
		close(ch)
	}()
	// read from channel as they come in until its closed
	for res := range ch {
		responses = append(responses, res)
	}

	return responses, nil
}
