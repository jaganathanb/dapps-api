package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jaganathanb/dapps-api/api/helper"
	"github.com/jaganathanb/dapps-api/config"
	service_errors "github.com/jaganathanb/dapps-api/pkg/service-errors"
	"github.com/jaganathanb/dapps-api/services"
)

type StreamerHandler struct {
	service *services.StreamerService
}

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
// @Success 200 {object} helper.BaseHttpResponse "Success"
// @Failure 400 {object} helper.BaseHttpResponse "Failed"
// @Router /v{version}/stream [get]
func (h *StreamerHandler) StreamData(c *gin.Context) {
	res, err := h.service.StreamData("", "")

	if err != nil {
		c.AbortWithStatusJSON(helper.TranslateErrorToStatusCode(err),
			helper.GenerateBaseResponseWithError(nil, false, helper.InternalError, err))
		return
	}

	if res == nil {
		c.AbortWithStatusJSON(helper.TranslateErrorToStatusCode(errors.New(service_errors.GstNotFound)),
			helper.GenerateBaseResponseWithError(nil, false, helper.InternalError, errors.New(service_errors.GstNotFound)))
		return
	}

	c.JSON(http.StatusOK, helper.GenerateBaseResponse(res, true, helper.Success))
}
