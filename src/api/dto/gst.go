package dto

import (
	"time"

	"github.com/jaganathanb/dapps-api/constants"
)

type CreateGstRequest struct {
	Gst
}

type CreateGstsRequest struct {
	BaseDto
	Gsts []Gst `json:"gsts" binding:"required"`
}

type GetGstResponse = Gst

type GetGstsResponse = []GetGstResponse

type GstFiledCount struct {
	GSTR1Count  int64 `json:"gstr1Count"`
	GSTR3BCount int64 `json:"gstr3bCount"`
	GSTR2Count  int64 `json:"gstr2Count"`
	GSTR9Count  int64 `json:"gstr9Count"`
	TotalGsts   int64 `json:"totalGsts"`
}

type UpdateGstReturnStatusRequest struct {
	BaseDto
	Gstin      string                    `json:"gstin" binding:"required,gstin"`
	ReturnType constants.GstReturnType   `json:"returnType"`
	Status     constants.GstReturnStatus `json:"status"`
}

type UpdateGstLockStatusRequest struct {
	BaseDto
	Gstin  string `json:"gstin"`
	Locked bool   `json:"locked"`
}

type RemoveGstRequest struct {
	BaseDto
	Gstin string `json:"gstin" binding:"required,gstin"`
}

type Gst struct {
	Sno              string           `json:"sno"`
	Fno              string           `json:"fno"`
	Gstin            string           `json:"gstin"`
	Name             string           `json:"name"`
	TradeName        string           `json:"tradeName"`
	Email            string           `json:"email"`
	RegistrationDate string           `json:"registrationDate"`
	Type             string           `json:"type"`
	LastUpdateDate   time.Time        `json:"lastUpdateDate"`
	Locked           bool             `json:"locked"`
	MobileNumber     string           `json:"mobileNumber"`
	Username         string           `json:"username"`
	Password         string           `json:"password"`
	GstStatuses      []GstStatus      `json:"gstStatuses"`
	PermenantAddress PermenantAddress `json:"permenantAddress"`
}

type PermenantAddress struct {
	Street   string `json:"street"`
	Locality string `json:"locality"`
	DoorNo   string `json:"doorNo"`
	State    string `json:"state"`
	Pincode  string `json:"pincode"`
	District string `json:"district"`
	City     string `json:"city"`
	LandMark string `json:"landMark"`
}

type GstStatus struct {
	Valid          string                    `json:"valid"`
	ModeOfFiling   string                    `json:"modeOfFiling"`
	LastFiledDate  string                    `json:"lastFiledDate"`
	ReturnType     constants.GstReturnType   `json:"returnType"`
	ReturnPeriod   string                    `json:"returnPeriod"`
	TaxPrd         string                    `json:"taxp"`
	Arn            string                    `json:"arn"`
	Status         constants.GstReturnStatus `json:"status"`
	Notes          string                    `json:"notes"`
	PendingReturns []string                  `json:"pendingReturns"`
}
