package models

import (
	"time"

	"github.com/jaganathanb/dapps-api/constants"
)

type Gst struct {
	BaseModel
	Gstin            string      `gorm:"type:string;size:30;not null,unique"`
	TradeName        string      `gorm:"type:string;size:64;not null"`
	RegistrationDate time.Time   `gorm:"type:TIMESTAMP;default:CURRENT_TIMESTAMP;not null"`
	Locked           bool        `gorm:"type:bool;default:false"`
	Address          string      `gorm:"type:string;size:128;null"`
	MobileNumber     string      `gorm:"type:string;size:10;null;default:null"`
	GstStatuses      []GstStatus `gorm:"foreignKey:Gstin;references:Gstin"`
}

type GstStatus struct {
	BaseModel
	GstRType       constants.GstReturnType   `gorm:"type:uint"`
	Status         constants.GstReturnStatus `gorm:"type:uint"`
	FiledDate      time.Time                 `gorm:"type:TIMESTAMP;default:CURRENT_TIMESTAMP;not null"`
	PendingReturns string
	TaxPeriod      string
	Notes          string
	Gstin          string `gorm:"type:string;size:30;not null,unique"`
}
