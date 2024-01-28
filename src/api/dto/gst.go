package dto

import (
	"time"

	"github.com/jaganathanb/dapps-api/constants"
)

type CreateGstRequest struct {
	Gstin            string      `json:"gstin" binding:"required,max=15"`
	TradeName        string      `json:"tradeName"`
	RegistrationDate time.Time   `json:"registrationDate"`
	Locked           bool        `json:"locked"`
	MobileNumber     string      `json:"mobileNumber" binding:"max=10"`
	Address          string      `json:"address" binding:"max=128"`
	GstStatuses      []GstStatus `json:"gstStatuses"`
}

type CreateGstsRequest struct {
	Gstins []string `json:"gstins" binding:"required"`
}

type GetGstResponse = CreateGstRequest

type GetGstsResponse = []GetGstResponse

type UpdateGstReturnStatusRequest struct {
	Gstin       string      `json:"gstin" binding:"required,max=15"`
	GstStatuses []GstStatus `json:"gstStatuses"`
}

type LockGstRequest struct {
	Gstin  string `json:"gstin" binding:"required,max=15"`
	Locked bool   `json:"locked"`
}

type RemoveGstRequest struct {
	Gstin string `json:"gstin" binding:"required,max=15"`
}

type GstStatus struct {
	GstRType       constants.GstReturnType
	Status         constants.GstReturnStatus
	FiledDate      time.Time
	PendingReturns string
	TaxPeriod      string
	Notes          string
}
