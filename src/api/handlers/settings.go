package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jaganathanb/dapps-api/api/dto"
	"github.com/jaganathanb/dapps-api/api/helper"
	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/services"
)

type SettingsHandler struct {
	service *services.SettingsService
}

func NewSettingsHandler(cfg *config.Config) *SettingsHandler {
	service := services.NewSettingsService(cfg)

	return &SettingsHandler{service: service}
}

// GetSettings godoc
// @Summary Get settings
// @Description Settings for GST Web
// @Tags Settings
// @Accept  json
// @Produce  json
// @Security AuthBearer
// @Param version path int true "Version" Enums(1, 2) default(1)
// @Success 200 {object} helper.BaseHttpResponse "Success"
// @Failure 400 {object} helper.BaseHttpResponse "Failed"
// @Router /v{version}/settings [get]
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	settings, err := h.service.GetSettings()

	if err != nil {
		c.AbortWithStatusJSON(helper.TranslateErrorToStatusCode(err),
			helper.GenerateBaseResponseWithError(nil, false, helper.InternalError, err))
		return
	}

	c.JSON(http.StatusOK, helper.GenerateBaseResponse(settings, true, 0))
}

// UpdateSettings godoc
// @Summary Update settings
// @Description Update settings for GST Web
// @Tags Settings
// @Accept  json
// @Produce  json
// @Security AuthBearer
// @Param version path int true "Version" Enums(1, 2) default(1)
// @Param Request body dto.SettingsPayload true "SettingsPayload"
// @Success 200 {object} helper.BaseHttpResponse "Success"
// @Failure 400 {object} helper.BaseHttpResponse "Failed"
// @Router /v{version}/settings [put]
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	req := new(dto.SettingsPayload)
	err := c.ShouldBindJSON(&req)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			helper.GenerateBaseResponseWithValidationError(nil, false, helper.ValidationError, err))
		return
	}

	header, ok := GetHeaderValues(c)

	if !ok {
		return
	}

	req.ModifiedBy = header.DappsUserId
	req.CreatedBy = header.DappsUserId
	settings, err := h.service.UpdateSettings(req)

	if err != nil {
		c.AbortWithStatusJSON(helper.TranslateErrorToStatusCode(err),
			helper.GenerateBaseResponseWithError(nil, false, helper.InternalError, err))
		return
	}

	c.JSON(http.StatusOK, helper.GenerateBaseResponse(settings, true, 0))
}
