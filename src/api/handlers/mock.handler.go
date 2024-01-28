package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jaganathanb/dapps-api/api/helper"
	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/services"
)

type MockHandler struct {
	service *services.MockService
}

func NewMockHandler(cfg *config.Config) *MockHandler {
	service := services.NewMockService(cfg)

	return &MockHandler{service: service}
}

// Mock godoc
// @Summary Mock data
// @Description Mock data endpoint
// @Tags Mock
// @Accept  json
// @Produce  json
// @Param version path int true "Version" Enums(1, 2) default(1)
// @Param filename path string true "File name"
// @Param prop path string true "Property name"
// @Success 200 {object} helper.BaseHttpResponse "Success"
// @Failure 400 {object} helper.BaseHttpResponse "Failed"
// @Router /v{version}/mocks/{filename}/{prop} [get]
func (h *MockHandler) GetMockData(c *gin.Context) {
	fileName := c.Params.ByName("filename")
	prop := c.Params.ByName("prop")

	res, err := h.service.GetMockData(fmt.Sprintf("%s.json", fileName), prop)

	if err != nil {
		c.AbortWithStatusJSON(helper.TranslateErrorToStatusCode(err),
			helper.GenerateBaseResponseWithError(nil, false, helper.InternalError, err))
		return
	}

	c.JSON(http.StatusOK, helper.GenerateBaseResponse(res, true, helper.Success))
}
