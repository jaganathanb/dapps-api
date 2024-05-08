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
	service         *services.GstService
	scrapperService *services.ScrapperService
}

func NewGstsHandler(cfg *config.Config) *GstsHandler {
	service := services.NewGstService(cfg)
	scrapperService := services.NewScrapperService(cfg)

	return &GstsHandler{service: service, scrapperService: scrapperService}
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

// CreateGsts godoc
// @Summary Creates GSTs
// @Description Create GST entries into the system
// @Tags GSTs
// @Accept  json
// @Produce  json
// @Security AuthBearer
// @Param version path int true "Version" Enums(1, 2) default(1)
// @Param Request body dto.CreateGstsRequest true "CreateGstsRequest"
// @Success 201 {object} helper.BaseHttpResponse "Success"
// @Failure 400 {object} helper.BaseHttpResponse "Failed"
// @Failure 409 {object} helper.BaseHttpResponse "Failed"
// @Router /v{version}/gsts [post]
func (h *GstsHandler) CreateGsts(c *gin.Context) {
	req := new(dto.CreateGstsRequest)
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

	req.CreatedBy = header.DappsUserId
	msg, err := h.service.CreateGsts(req)
	if err != nil {
		c.AbortWithStatusJSON(helper.TranslateErrorToStatusCode(err),
			helper.GenerateBaseResponseWithValidationError(nil, false, helper.ValidationError, err))
		return
	}

	c.JSON(http.StatusCreated, helper.GenerateBaseResponse(msg, true, helper.Success))
}

// UpdateGstStatus godoc
// @Summary Updates GST statuses
// @Description Updates the statuses of the GST entry into the system
// @Tags GSTs
// @Accept  json
// @Produce  json
// @Security AuthBearer
// @Param version path int true "Version" Enums(1, 2) default(1)
// @Param gstin path string true "Gstin"
// @Param Request body dto.UpdateGstReturnStatusRequest true "UpdateGstStatus"
// @Success 201 {object} helper.BaseHttpResponse "Success"
// @Failure 400 {object} helper.BaseHttpResponse "Failed"
// @Failure 409 {object} helper.BaseHttpResponse "Failed"
// @Router /v{version}/gsts/{gstin}/return-status [put]
func (h *GstsHandler) UpdateGstStatus(c *gin.Context) {
	gstin := c.Params.ByName("gstin")
	if gstin == "" {
		c.AbortWithStatusJSON(http.StatusNotFound,
			helper.GenerateBaseResponse(nil, false, helper.ValidationError))
		return
	}

	req := new(dto.UpdateGstReturnStatusRequest)
	req.Gstin = gstin
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
	ok, err = h.service.UpdateGstStatus(req)
	if err != nil {
		c.AbortWithStatusJSON(helper.TranslateErrorToStatusCode(err),
			helper.GenerateBaseResponseWithError(nil, false, helper.InternalError, err))
		return
	}

	c.JSON(http.StatusCreated, helper.GenerateBaseResponse(ok, true, helper.Success))
}

// LockGstById godoc
// @Summary Updates GST lock status
// @Description Updates the lock status of GST in system
// @Tags GSTs
// @Accept  json
// @Produce  json
// @Security AuthBearer
// @Param version path int true "Version" Enums(1, 2) default(1)
// @Param gstin path string true "Gstin"
// @Param Request body dto.UpdateGstReturnStatusRequest true "Locked"
// @Success 201 {object} helper.BaseHttpResponse "Success"
// @Failure 400 {object} helper.BaseHttpResponse "Failed"
// @Failure 409 {object} helper.BaseHttpResponse "Failed"
// @Router /v{version}/gsts/{gstin}/lock [put]
func (h *GstsHandler) LockGstById(c *gin.Context) {
	gstin := c.Params.ByName("gstin")
	if gstin == "" {
		c.AbortWithStatusJSON(http.StatusNotFound,
			helper.GenerateBaseResponse(nil, false, helper.ValidationError))
		return
	}

	req := new(dto.UpdateGstLockStatusRequest)
	req.Gstin = gstin

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

	ok, err = h.service.LockGstById(req)
	if err != nil {
		c.AbortWithStatusJSON(helper.TranslateErrorToStatusCode(err),
			helper.GenerateBaseResponseWithError(nil, false, helper.InternalError, err))
		return
	}

	c.JSON(http.StatusCreated, helper.GenerateBaseResponse(ok, true, helper.Success))
}

// DeleteGstById godoc
// @Summary Deletes GST by id
// @Description Deletes the given GST from system
// @Tags GSTs
// @Accept  json
// @Produce  json
// @Security AuthBearer
// @Param version path int true "Version" Enums(1, 2) default(1)
// @Param gstin path string true "Gstin"
// @Success 200 {object} helper.BaseHttpResponse "Success"
// @Failure 400 {object} helper.BaseHttpResponse "Failed"
// @Failure 409 {object} helper.BaseHttpResponse "Failed"
// @Router /v{version}/gsts/{gstin} [delete]
func (h *GstsHandler) DeleteGstById(c *gin.Context) {
	gstin := c.Params.ByName("gstin")
	if gstin == "" {
		c.AbortWithStatusJSON(http.StatusNotFound,
			helper.GenerateBaseResponse(nil, false, helper.ValidationError))
		return
	}

	req := new(dto.RemoveGstRequest)
	req.Gstin = gstin

	header, ok := GetHeaderValues(c)
	if !ok {
		return
	}

	req.DeletedBy = header.DappsUserId

	ok, err := h.service.DeleteGstById(req)
	if err != nil {
		c.AbortWithStatusJSON(helper.TranslateErrorToStatusCode(err),
			helper.GenerateBaseResponseWithError(nil, false, helper.InternalError, err))
		return
	}

	c.JSON(http.StatusOK, helper.GenerateBaseResponse(ok, true, helper.Success))
}

// GetGstStatistics godoc
// @Summary Gets GST statistics
// @Description Gets no of gsts filed for all the return types for current last tax period
// @Tags GSTs
// @Accept  json
// @Produce  json
// @Security AuthBearer
// @Param version path int true "Version" Enums(1, 2) default(1)
// @Success 200 {object} helper.BaseHttpResponse{result=dto.GstFiledCount} "Success"
// @Failure 400 {object} helper.BaseHttpResponse "Failed"
// @Failure 409 {object} helper.BaseHttpResponse "Failed"
// @Router /v{version}/gsts/statistics [get]
func (h *GstsHandler) GetGstStatistics(c *gin.Context) {
	statistics, err := h.service.GetGstStatistics()
	if err != nil {
		c.AbortWithStatusJSON(helper.TranslateErrorToStatusCode(err),
			helper.GenerateBaseResponseWithError(nil, false, helper.InternalError, err))
		return
	}

	c.JSON(http.StatusOK, helper.GenerateBaseResponse(statistics, true, helper.Success))
}

// RefreshGstReturns godoc
// @Summary Refreshes GST returns
// @Description Refreshes the gst returns who are in state `EntryDone` or there is no returns yet
// @Tags GSTs
// @Accept  json
// @Produce  json
// @Security AuthBearer
// @Param version path int true "Version" Enums(1, 2) default(1)
// @Success 200 {object} helper.BaseHttpResponse "Success"
// @Failure 400 {object} helper.BaseHttpResponse "Failed"
// @Failure 409 {object} helper.BaseHttpResponse "Failed"
// @Router /v{version}/gsts/refresh-returns [get]
func (h *GstsHandler) RefreshGstReturns(c *gin.Context) {
	err := h.service.RefreshGstReturns()

	if err != nil {
		c.AbortWithStatusJSON(helper.TranslateErrorToStatusCode(err),
			helper.GenerateBaseResponseWithError(nil, false, helper.InternalError, err))
		return
	}

	c.JSON(http.StatusOK, helper.GenerateBaseResponse(nil, true, helper.Success))
}
