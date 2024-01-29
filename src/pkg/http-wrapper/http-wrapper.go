package httpwrapper

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/jaganathanb/dapps-api/api/dto"
	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/pkg/logging"
)

var log = logging.NewLogger(config.GetConfig())
var client = http.Client{}

func makeCall[T any](req *http.Request, ch chan<- dto.HttpResonseWrapper[T], wg *sync.WaitGroup) {
	defer wg.Done()

	resp, err := client.Do(req)
	if err != nil {
		log.Error(logging.Category(logging.ExternalService), logging.SubCategory(logging.RequestResponse), err.Error(), nil)

		ch <- dto.HttpResonseWrapper[T]{Resonse: nil, Error: err}
		return
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if resp.StatusCode == 201 || resp.StatusCode == 200 {
		if err != nil {
			log.Error(logging.Category(logging.ExternalService), logging.SubCategory(logging.RequestResponse), err.Error(), nil)

			ch <- dto.HttpResonseWrapper[T]{Resonse: nil, Error: err}
			return
		}

		var res *T
		err = json.Unmarshal(b, &res)

		if err != nil {
			log.Error(logging.Category(logging.ExternalService), logging.SubCategory(logging.RequestResponse), err.Error(), nil)

			ch <- dto.HttpResonseWrapper[T]{Resonse: nil, Error: err}
			return
		}

		ch <- dto.HttpResonseWrapper[T]{Resonse: res, Error: nil}
	} else {
		ch <- dto.HttpResonseWrapper[T]{Resonse: nil, Error: errors.New(fmt.Sprintf("HTTP Error %d. Error: %s", resp.StatusCode, b))}
	}
}

func AsyncHTTP[T any](reqs []http.Request) ([]dto.HttpResonseWrapper[T], error) {
	ch := make(chan dto.HttpResonseWrapper[T])
	var responses []dto.HttpResonseWrapper[T]
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
