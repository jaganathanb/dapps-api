package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jaganathanb/dapps-api/api/dto"
	"github.com/jaganathanb/dapps-api/api/helper"
	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/services"
)

type GstsHandler struct {
	service *services.GstService
}

func NewGstsHandler(cfg *config.Config) *GstsHandler {
	service := services.NewGstService(cfg)

	return &GstsHandler{service: service}
}

// CreateGST godoc
// @Summary Creates GST
// @Description Create a GST entry into the system
// @Tags GSTs
// @Accept  json
// @Produce  json
// @Security AuthBearer
// @Param version path int true "Version"
// @Param Request body dto.CreateGSTRequest true "CreateGSTRequest"
// @Success 201 {object} helper.BaseHttpResponse "Success"
// @Failure 400 {object} helper.BaseHttpResponse "Failed"
// @Failure 409 {object} helper.BaseHttpResponse "Failed"
// @Router /v{version}/gsts [post]
func (h *GstsHandler) CreateGST(c *gin.Context) {
	req := new(dto.CreateGSTRequest)
	err := c.ShouldBindJSON(&req)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			helper.GenerateBaseResponseWithValidationError(nil, false, helper.ValidationError, err))
		return
	}
	err = h.service.CreateGST(req)
	if err != nil {
		c.AbortWithStatusJSON(helper.TranslateErrorToStatusCode(err),
			helper.GenerateBaseResponseWithError(nil, false, helper.InternalError, err))
		return
	}

	c.JSON(http.StatusCreated, helper.GenerateBaseResponse(nil, true, helper.Success))
}
