package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jaganathanb/dapps-api/api/dto"
	"github.com/jaganathanb/dapps-api/api/helper"
	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/services"
)

type NotificationsHandler struct {
	service *services.NotificationsService
}

func NewNotificationsHandler(cfg *config.Config) *NotificationsHandler {
	service := services.NewNotificationsService(cfg)

	return &NotificationsHandler{service: service}
}

// GetNotifications godoc
// @Summary Get notifications
// @Description Notifications for GST Web
// @Tags Notifications
// @Accept  json
// @Produce  json
// @Security AuthBearer
// @Param version path int true "Version" Enums(1, 2) default(1)
// @Success 200 {object} helper.BaseHttpResponse "Success"
// @Failure 400 {object} helper.BaseHttpResponse "Failed"
// @Router /v{version}/notifications [get]
func (h *NotificationsHandler) GetNotifications(c *gin.Context) {
	notifications, err := h.service.GetNotifications()

	if err != nil {
		c.AbortWithStatusJSON(helper.TranslateErrorToStatusCode(err),
			helper.GenerateBaseResponseWithError(nil, false, helper.InternalError, err))
		return
	}

	c.JSON(http.StatusOK, helper.GenerateBaseResponse(notifications, true, 0))
}

// AddNotifications godoc
// @Summary Add notifications
// @Description Add notifications for GST Web
// @Tags Notifications
// @Accept  json
// @Produce  json
// @Security AuthBearer
// @Param version path int true "Version" Enums(1, 2) default(1)
// @Param Request body dto.NotificationsPayload true "NotificationsPayload"
// @Success 201 {object} helper.BaseHttpResponse "Success"
// @Failure 400 {object} helper.BaseHttpResponse "Failed"
// @Router /v{version}/notifications [post]
func (h *NotificationsHandler) AddNotifications(c *gin.Context) {
	req := new(dto.NotificationsPayload)
	err := c.ShouldBindJSON(&req)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			helper.GenerateBaseResponseWithValidationError(nil, false, helper.ValidationError, err))
		return
	}
	notifications, err := h.service.AddNotification(req)

	if err != nil {
		c.AbortWithStatusJSON(helper.TranslateErrorToStatusCode(err),
			helper.GenerateBaseResponseWithError(nil, false, helper.InternalError, err))
		return
	}

	c.JSON(http.StatusOK, helper.GenerateBaseResponse(notifications, true, 0))
}

// UpdateNotifications godoc
// @Summary Update notifications
// @Description Update notifications for GST Web
// @Tags Notifications
// @Accept  json
// @Produce  json
// @Security AuthBearer
// @Param version path int true "Version" Enums(1, 2) default(1)
// @Param Request body dto.NotificationsPayload true "NotificationsPayload"
// @Success 200 {object} helper.BaseHttpResponse "Success"
// @Failure 400 {object} helper.BaseHttpResponse "Failed"
// @Router /v{version}/notifications [put]
func (h *NotificationsHandler) UpdateNotifications(c *gin.Context) {
	req := new(dto.NotificationsPayload)
	err := c.ShouldBindJSON(&req)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			helper.GenerateBaseResponseWithValidationError(nil, false, helper.ValidationError, err))
		return
	}
	notifications, err := h.service.UpdateNotifications(req)

	if err != nil {
		c.AbortWithStatusJSON(helper.TranslateErrorToStatusCode(err),
			helper.GenerateBaseResponseWithError(nil, false, helper.InternalError, err))
		return
	}

	c.JSON(http.StatusOK, helper.GenerateBaseResponse(notifications, true, 0))
}
