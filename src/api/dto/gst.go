package dto

import (
	"time"

	"github.com/jaganathanb/dapps-api/constants"
)

type CreateGSTRequest struct {
	Gstin            string      `json:"gstin" binding:"required,max=15"`
	TradeName        string      `json:"tradeName"`
	RegistrationDate time.Time   `json:"registrationDate"`
	Locked           bool        `json:"locked"`
	MobileNumber     string      `json:"mobileNumber" binding:"max=10"`
	Address          string      `json:"address" binding:"max=128"`
	GSTStatuses      []GSTStatus `json:"gstStatuses"`
}

type UpdateGSTReturnStatusRequest struct {
	Gstin       string      `json:"gstin" binding:"required,max=15"`
	GSTStatuses []GSTStatus `json:"gstStatuses"`
}

type LockGSTRequest struct {
	Gstin  string `json:"gstin" binding:"required,max=15"`
	Locked bool   `json:"locked"`
}

type RemoveGSTRequest struct {
	Gstin string `json:"gstin" binding:"required,max=15"`
}

type GSTStatus struct {
	GstRType       constants.GstReturnType
	Status         constants.GstReturnStatus
	FiledDate      time.Time
	PendingReturns string
	TaxPeriod      string
	Notes          string
}
