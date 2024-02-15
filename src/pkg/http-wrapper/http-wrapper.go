package httpwrapper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sync"

	"github.com/jaganathanb/dapps-api/api/dto"
)

// HTTPClient represents the HTTP client wrapper
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient creates a new HTTPClient instance
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &http.Client{},
	}
}

// MakeRequests sends concurrent requests and returns a channel to receive responses
func (c *HTTPClient) MakeRequests(requests []dto.HttpRequestConfig) []dto.HttpResponseWrapper {
	responses := c.doRequests(requests)

	resps := []dto.HttpResponseWrapper{}
	// Process responses
	for response := range responses {
		if response.Err != nil {
			fmt.Println("Error:", response.Err)
			resps = append(resps, response)
			continue
		}

		resps = append(resps, response)
	}

	return resps
}

func (c *HTTPClient) doRequests(requests []dto.HttpRequestConfig) <-chan dto.HttpResponseWrapper {
	var wg sync.WaitGroup

	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)

	maxConcurrency := 2 * numCPU

	resultChan := make(chan dto.HttpResponseWrapper, len(requests))
	requestChan := make(chan dto.HttpRequestConfig, len(requests))

	for i := 0; i < maxConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for req := range requestChan {
				resultChan <- c.doRequest(req)
			}
		}()
	}

	go func() {
		defer close(requestChan)
		for _, req := range requests {
			requestChan <- req
		}
	}()

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	return resultChan
}

// doRequest sends an HTTP request based on the provided method
func (c *HTTPClient) doRequest(config dto.HttpRequestConfig) dto.HttpResponseWrapper {
	var bodyBytes []byte
	if config.Body != nil {
		var err error
		bodyBytes, err = json.Marshal(config.Body)
		if err != nil {
			return dto.HttpResponseWrapper{Err: err, RequestID: config.RequestID}
		}
	}

	req, err := http.NewRequest(config.Method, config.URL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return dto.HttpResponseWrapper{Err: err, RequestID: config.RequestID}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return dto.HttpResponseWrapper{Err: err, RequestID: config.RequestID}
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return dto.HttpResponseWrapper{Err: fmt.Errorf(resp.Status), RequestID: config.RequestID}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return dto.HttpResponseWrapper{Err: err, RequestID: config.RequestID}
	}

	var responseBody any
	if config.ResponseType != nil {
		err := json.Unmarshal(body, config.ResponseType)
		if err != nil {
			return dto.HttpResponseWrapper{Err: err, RequestID: config.RequestID}
		}
		responseBody = config.ResponseType
	} else {
		responseBody = body
	}

	return dto.HttpResponseWrapper{StatusCode: resp.StatusCode, RequestID: config.RequestID, Body: responseBody, ResponseType: config.ResponseType}
}
