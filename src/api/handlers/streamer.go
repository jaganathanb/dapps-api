package handlers

import (
	"io"

	"github.com/gin-gonic/gin"
	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/services"
)

type StreamerHandler struct {
	service *services.StreamerService
}

type ClientChan chan string

func NewStreamerHandler(cfg *config.Config) *StreamerHandler {
	service := services.NewStreamerService(cfg)

	return &StreamerHandler{service: service}
}

// Streamer godoc
// @Summary Stream data
// @Description Stream data endpoint
// @Tags Stream
// @Accept  json
// @Produce  json
// @Param version path int true "Version" Enums(1, 2) default(1)
// @Router /v{version}/stream [get]
func (h *StreamerHandler) StreamData(c *gin.Context) {
	v, ok := c.Get("clientChan")
	if !ok {
		return
	}
	clientChan, ok := v.(ClientChan)
	if !ok {
		return
	}
	c.Stream(func(w io.Writer) bool {
		// Stream message to client from message channel
		if msg, ok := <-clientChan; ok {
			c.SSEvent("message", msg)
			return true
		}
		return false
	})
}

func (h *StreamerHandler) ServeStream() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Initialize client channel
		clientChan := make(ClientChan)

		// Send new connection to event server
		h.service.AddClient(clientChan)

		defer func() {
			// Send closed connection to event server
			h.service.RemoveClient(clientChan)
		}()

		c.Set("clientChan", clientChan)

		c.Next()
	}
}
