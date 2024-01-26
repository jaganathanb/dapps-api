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

// GetGsts godoc
// @Summary Gets GST
// @Description Gets all available GSTs from the system
// @Tags GSTs
// @Accept json
// @produces json
// @Param version path int true "Version" Enums(1, 2) default(1)
// @Param Request body dto.PaginationInputWithFilter true "Request"
// @Success 200 {object} helper.BaseHttpResponse{result=dto.PagedList[dto.GetGstResponse]} "GetGst response"
// @Failure 400 {object} helper.BaseHttpResponse "Bad request"
// @Router /v{version}/gsts/page [post]
// @Security AuthBearer
func (h *GstsHandler) GetGsts(c *gin.Context) {
	GetByFilter(c, h.service.GetByFilter)
}

// CreateGst godoc
// @Summary Creates GST
// @Description Create a GST entry into the system
// @Tags GSTs
// @Accept  json
// @Produce  json
// @Security AuthBearer
// @Param version path int true "Version" Enums(1, 2) default(1)
// @Param Request body dto.CreateGSTRequest true "CreateGSTRequest"
// @Success 201 {object} helper.BaseHttpResponse "Success"
// @Failure 400 {object} helper.BaseHttpResponse "Failed"
// @Failure 409 {object} helper.BaseHttpResponse "Failed"
// @Router /v{version}/gsts [post]
func (h *GstsHandler) CreateGst(c *gin.Context) {
	req := new(dto.CreateGSTRequest)
	err := c.ShouldBindJSON(&req)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			helper.GenerateBaseResponseWithValidationError(nil, false, helper.ValidationError, err))
		return
	}
	err = h.service.CreateGst(req)
	if err != nil {
		c.AbortWithStatusJSON(helper.TranslateErrorToStatusCode(err),
			helper.GenerateBaseResponseWithError(nil, false, helper.InternalError, err))
		return
	}

	c.JSON(http.StatusCreated, helper.GenerateBaseResponse(nil, true, helper.Success))
}

// UpdateGstStatuses godoc
// @Summary Updates GST statuses
// @Description Updates the statuses of the GST entry into the system
// @Tags GSTs
// @Accept  json
// @Produce  json
// @Security AuthBearer
// @Param version path int true "Version" Enums(1, 2) default(1)
// @Param gstin path string true "Gstin"
// @Param Request body dto.UpdateGSTReturnStatusRequest true "UpdateGstStatuses"
// @Success 201 {object} helper.BaseHttpResponse "Success"
// @Failure 400 {object} helper.BaseHttpResponse "Failed"
// @Failure 409 {object} helper.BaseHttpResponse "Failed"
// @Router /v{version}/gsts/{gstin}/statuses [put]
func (h *GstsHandler) UpdateGstStatuses(c *gin.Context) {
	gstin := c.Params.ByName("gstin")
	if gstin == "" {
		c.AbortWithStatusJSON(http.StatusNotFound,
			helper.GenerateBaseResponse(nil, false, helper.ValidationError))
		return
	}

	req := new(dto.UpdateGSTReturnStatusRequest)
	req.Gstin = gstin
	err := c.ShouldBindJSON(&req)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			helper.GenerateBaseResponseWithValidationError(nil, false, helper.ValidationError, err))
		return
	}
	err = h.service.UpdateGstStatuses(req)
	if err != nil {
		c.AbortWithStatusJSON(helper.TranslateErrorToStatusCode(err),
			helper.GenerateBaseResponseWithError(nil, false, helper.InternalError, err))
		return
	}

	c.JSON(http.StatusCreated, helper.GenerateBaseResponse(nil, true, helper.Success))
}
