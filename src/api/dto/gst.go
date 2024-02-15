package dto

import (
	"time"

	"github.com/jaganathanb/dapps-api/constants"
)

type CreateGstRequest struct {
	Gstin            string      `json:"gstin" binding:"required,len=15"`
	TradeName        string      `json:"tradeName"`
	RegistrationDate time.Time   `json:"registrationDate"`
	Locked           bool        `json:"locked"`
	MobileNumber     string      `json:"mobileNumber" binding:"max=10"`
	Address          string      `json:"address" binding:"max=128"`
	GstStatuses      []GstStatus `json:"gstStatuses"`
}

type CreateGstsRequest struct {
	Gstins []string `json:"gstins" binding:"required,gstins"`
}

type GetGstResponse = CreateGstRequest

type GetGstsResponse = []GetGstResponse

type UpdateGstReturnStatusRequest struct {
	Gstin       string      `json:"gstin" binding:"required,len=15,gstin"`
	GstStatuses []GstStatus `json:"gstStatuses"`
}

type LockGstRequest struct {
	Gstin  string `json:"gstin" binding:"required,max=15"`
	Locked bool   `json:"locked"`
}

type RemoveGstRequest struct {
	Gstin string `json:"gstin" binding:"required,max=15"`
}

type Gst struct {
	Gstin            string      `gorm:"type:string;size:30;not null,unique"`
	TradeName        string      `gorm:"type:string;size:64;not null"`
	RegistrationDate time.Time   `gorm:"type:TIMESTAMP;default:CURRENT_TIMESTAMP;not null"`
	Locked           bool        `gorm:"type:bool;default:false"`
	Address          string      `gorm:"type:string;size:128;null"`
	MobileNumber     string      `gorm:"type:string;size:10;null;default:null"`
	GstStatuses      []GstStatus `gorm:"foreignKey:Gstin;references:Gstin"`
}

type GstStatus struct {
	GstRType       constants.GstReturnType
	Status         constants.GstReturnStatus
	FiledDate      string
	PendingReturns []string
	TaxPeriod      string
	Notes          string
}
