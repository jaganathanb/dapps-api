package services

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/pkg/logging"
)

type StreamerService struct {
	logger     logging.Logger
	cfg        *config.Config
	httpClient http.Client
	streamer   *Streamer
}

type Streamer struct {
	// Events are pushed to this channel by the main events-gathering routine
	Message chan string

	// New client connections
	NewClients chan chan string

	// Closed client connections
	ClosedClients chan chan string

	// Total client connections
	TotalClients map[chan string]bool

	Logger logging.Logger
}

var streamerService *StreamerService
var streamerServiceOnce sync.Once

func NewStreamerService(cfg *config.Config) *StreamerService {
	streamerServiceOnce.Do(func() {
		logger := logging.NewLogger(cfg)
		client := http.Client{}
		streamer := getNewServer(logger)

		streamerService = &StreamerService{logger: logger, cfg: cfg, httpClient: client, streamer: streamer}
	})

	return streamerService
}

func (s *StreamerService) StreamData(message string) {
	s.streamer.Message <- fmt.Sprintf("%s|%s", message, time.Now())
}

func (s *StreamerService) AddClient(client chan string) {
	s.streamer.NewClients <- client
}

func (s *StreamerService) RemoveClient(client chan string) {
	s.streamer.ClosedClients <- client
}

func getNewServer(logger logging.Logger) (event *Streamer) {
	event = &Streamer{
		Message:       make(chan string),
		NewClients:    make(chan chan string),
		ClosedClients: make(chan chan string),
		TotalClients:  make(map[chan string]bool),
		Logger:        logger,
	}

	go event.listen()

	return
}

func (streamer *Streamer) listen() {
	for {
		select {
		// Add new available client
		case client := <-streamer.NewClients:
			streamer.TotalClients[client] = true
			streamer.Logger.Infof("Client added. %d registered clients", len(streamer.TotalClients))

		// Remove closed client
		case client := <-streamer.ClosedClients:
			delete(streamer.TotalClients, client)
			close(client)
			streamer.Logger.Infof("Removed client. %d registered clients", len(streamer.TotalClients))

		// Broadcast message to client
		case eventMsg := <-streamer.Message:
			for clientMessageChan := range streamer.TotalClients {
				clientMessageChan <- eventMsg
			}
		}
	}
}
