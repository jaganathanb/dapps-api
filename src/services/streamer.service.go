package services

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/jaganathanb/dapps-api/api/dto"
	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/constants"
	"github.com/jaganathanb/dapps-api/pkg/logging"
)

type StreamerService struct {
	logger              logging.Logger
	cfg                 *config.Config
	httpClient          http.Client
	streamer            *Streamer
	notificationService *NotificationsService
}

type StreamMessage struct {
	Message     string                            `json:"message"`
	MessageType constants.NotificationMessageType `json:"messageType"`
	Title       string                            `json:"title"`
	Code        string                            `json:"code"`
	UserId      int                               `json:"userId"`
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

		streamerService = &StreamerService{logger: logger, cfg: cfg, httpClient: client, streamer: streamer, notificationService: NewNotificationsService(cfg)}
	})

	return streamerService
}

func (s *StreamerService) StreamData(message StreamMessage) {
	msg, err := json.Marshal(message)

	if err != nil {
		s.logger.Error(logging.IO, logging.Api, err.Error(), nil)
		return
	}

	if message.Code == "NOTIFICATION" {
		s.notificationService.AddNotification(&dto.NotificationsPayload{Message: message.Message, MessageType: message.MessageType, Title: message.Title, UserId: message.UserId, BaseDto: dto.BaseDto{CreatedBy: message.UserId}})
	}

	s.streamer.Message <- string(msg)
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
