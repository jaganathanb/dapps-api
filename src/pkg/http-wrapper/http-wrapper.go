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
	"github.com/jaganathanb/dapps-api/data/models"
	"github.com/jaganathanb/dapps-api/pkg/logging"
)

type GST = models.Gst

type HttpResult interface {
	GST
}

var log = logging.NewLogger(config.GetConfig())
var client = http.Client{}

func makeCall[T HttpResult](req *http.Request, ch chan<- dto.HttpResonseWrapper[T], wg *sync.WaitGroup) {
	defer wg.Done()

	resp, err := client.Do(req)
	if err != nil {
		log.Error(logging.Category(logging.ExternalService), logging.SubCategory(logging.RequestResponse), err.Error(), nil)

		ch <- dto.HttpResonseWrapper[T]{Data: nil, Error: err}
		return
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if resp.StatusCode == 201 || resp.StatusCode == 200 {
		if err != nil {
			log.Error(logging.Category(logging.ExternalService), logging.SubCategory(logging.RequestResponse), err.Error(), nil)

			ch <- dto.HttpResonseWrapper[T]{Data: nil, Error: err}
			return
		}

		res := new(dto.HttpResponseResult[T])
		err := json.Unmarshal(b, &res)

		if err != nil {
			log.Error(logging.Category(logging.ExternalService), logging.SubCategory(logging.RequestResponse), err.Error(), nil)

			ch <- dto.HttpResonseWrapper[T]{Data: nil, Error: err}
			return
		}

		// result := res["result"].(map[string]interface{})
		// resData, ok := result.(*T)

		// if ok {
		// 	fmt.Printf("%v", resData)
		// }

		//t := new(T)
		//obj := any(t).(*T)

		ch <- dto.HttpResonseWrapper[T]{Data: &dto.HttpResponseResult[T]{Result: res.Result}}
	} else {
		ch <- dto.HttpResonseWrapper[T]{Data: nil, Error: errors.New(fmt.Sprintf("HTTP Error %d for the url %s. Error: %s", resp.StatusCode, req.URL.String(), b))}
	}
}

func AsyncHTTP[T HttpResult](reqs []http.Request) ([]dto.HttpResonseWrapper[T], error) {
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
