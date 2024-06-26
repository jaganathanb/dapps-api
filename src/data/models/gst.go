package models

import (
	"time"

	"github.com/jaganathanb/dapps-api/constants"
	sqlite_custom_type "github.com/jaganathanb/dapps-api/pkg/sqlite-custom-type"
)

type Gst struct {
	BaseModel
	Sno              string                            `json:"sno"`
	Fno              string                            `json:"fno"`
	Gstin            string                            `gorm:"primaryKey;type:string;size:30;index;unique"`
	Name             string                            `json:"lgnm"`
	Tradename        string                            `json:"tradeNam"`
	RegistrationDate string                            `json:"rgdt"`
	Center           string                            `json:"ctj"`
	State            string                            `json:"stj"`
	CenterCd         string                            `json:"center_cd"`
	StateCd          string                            `json:"state_cd"`
	Constitution     string                            `json:"ctb"`
	Type             string                            `json:"dty"`
	Status           string                            `json:"sts"`
	Username         string                            `json:"username"`
	Password         string                            `json:"password"`
	LastUpdateDate   time.Time                         `json:"lastUpdateDate"`
	CancellationDate time.Time                         `json:"cancellationDate"`
	Nature           sqlite_custom_type.SqliteStrArray `json:"nba,omitempty;type:text[]"`
	EinvoiceStatus   string                            `json:"einvoiceStatus"`
	Adadr            []AdditionalAddress               `gorm:"foreignKey:Gstin;references:Gstin"`
	Locked           bool                              `gorm:"type:bool;default:false"`
	MobileNumber     string                            `gorm:"type:string;size:10;null;default:null"`
	Email            string                            `json:"email"`
	GstStatuses      []GstStatus                       `gorm:"foreignKey:Gstin;references:Gstin"`
	Pradr            PermenantAddress                  `gorm:"foreignKey:Gstin;references:Gstin"`
	Contacted        MobEmail                          `gorm:"foreignKey:Gstin;references:Gstin"`
}

type MobEmail struct {
	Gstin  string `json:"gstin"`
	MobNum int64  `json:"mobNum"`
	Email  string `json:"email"`
}

type AdditionalAddress struct {
	BaseModel
	Gstin string  `json:"gstin"`
	Addr  Address `gorm:"foreignKey:Id;references:Id"`
	Ntr   string  `json:"ntr"`
}

type Address struct {
	BaseModel
	Bnm        string `json:"bnm"`
	St         string `json:"st"`
	Loc        string `json:"loc"`
	Bno        string `json:"bno"`
	Dst        string `json:"dst"`
	Stcd       string `json:"stcd"`
	Pncd       string `json:"pncd"`
	Locality   string `json:"locality"`
	Geocodelvl string `json:"geocodelvl"`
	Lg         string `json:"lg"`
	Lt         string `json:"lt"`
	LandMark   string `json:"landMark"`
	City       string `json:"city"`
	Flno       string `json:"flno"`
}

type PermenantAddress struct {
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
	Valid          string                            `json:"valid"`
	Mof            string                            `json:"mof"`
	Dof            string                            `json:"dof"`
	Rtntype        constants.GstReturnType           `json:"rtntype"`
	RetPrd         string                            `json:"ret_prd"`
	TaxPrd         string                            `json:"taxp"`
	FinancialYear  string                            `json:"fy"`
	Arn            string                            `json:"arn"`
	Status         constants.GstReturnStatus         `json:"status"`
	Notes          string                            `json:"notes"`
	PendingReturns sqlite_custom_type.SqliteStrArray `json:"pending_returns,omitempty;type:text[]"`
	Gstin          string                            `gorm:"type:string,not null;size:30"`
}
