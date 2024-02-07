package models

import (
	"time"

	"github.com/jaganathanb/dapps-api/constants"
	sqlite_array "github.com/jaganathanb/dapps-api/pkg/sqlite-array-type"
)

type Gst struct {
	BaseModel
	Gstin            string                      `json:"gstin"`
	Name             string                      `json:"name"`
	Tradename        string                      `json:"tradename"`
	RegistrationDate string                      `json:"registrationDate"`
	Center           string                      `json:"center"`
	State            string                      `json:"state"`
	CenterCd         string                      `json:"center_cd"`
	StateCd          string                      `json:"state_cd"`
	Constitution     string                      `json:"constitution"`
	Type             string                      `json:"type"`
	Status           string                      `json:"status"`
	LastUpdateDate   time.Time                   `json:"lastUpdateDate"`
	CancellationDate time.Time                   `json:"cancellationDate"`
	Nature           sqlite_array.SqliteStrArray `json:"nature,omitempty;type:text[]"`
	EinvoiceStatus   string                      `json:"einvoiceStatus"`
	Adadr            sqlite_array.SqliteStrArray `json:"adadr,omitempty;type:text[]"`
	Locked           bool                        `gorm:"type:bool;default:false"`
	MobileNumber     string                      `gorm:"type:string;size:10;null;default:null"`
	GstStatuses      []GstStatus                 `gorm:"foreignKey:Gstin;references:Gstin"`
	Pradr            PAddress                    `gorm:"foreignKey:Gstin;references:Gstin"`
}

type PAddress struct {
	BaseModel
	Gstin      string `json:"gstin"`
	Bnm        string `json:"bnm"`
	St         string `json:"st"`
	Loc        string `json:"loc"`
	Bno        string `json:"bno"`
	Stcd       string `json:"stcd"`
	Flno       string `json:"flno"`
	Lt         string `json:"lt"`
	Lg         string `json:"lg"`
	Pncd       string `json:"pncd"`
	Ntr        string `json:"ntr"`
	District   string `json:"district"`
	City       string `json:"city"`
	Locality   string `json:"locality"`
	Geocodelvl string `json:"geocodelvl"`
	LandMark   string `json:"landMark"`
}

type GstStatus struct {
	BaseModel
	GstRType       constants.GstReturnType   `gorm:"type:uint"`
	Status         constants.GstReturnStatus `gorm:"type:uint"`
	FiledDate      time.Time                 `gorm:"type:TIMESTAMP;default:CURRENT_TIMESTAMP;not null"`
	PendingReturns string
	TaxPeriod      string
	Notes          string
	ArnNumber      string
	Gstin          string `gorm:"type:string;size:30;not null,unique"`
}
