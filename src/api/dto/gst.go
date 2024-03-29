package dto

import (
	"time"

	"github.com/jaganathanb/dapps-api/constants"
)

type CreateGstRequest struct {
	Gst
}

type CreateGstsRequest struct {
	Gsts []Gst `json:"gsts" binding:"required"`
}

type GetGstResponse = Gst

type GetGstsResponse = []GetGstResponse

type UpdateGstReturnStatusRequest struct {
	Gstin      string                    `json:"gstin" binding:"required,gstin"`
	ReturnType constants.GstReturnType   `json:"returnType"`
	Status     constants.GstReturnStatus `json:"status"`
}

type UpdateGstLockStatusRequest struct {
	Gstin  string `json:"gstin"`
	Locked bool   `json:"locked"`
}

type RemoveGstRequest struct {
	Gstin string `json:"gstin" binding:"required,max=15"`
}

type Gst struct {
	Gstin            string           `json:"gstin"`
	Name             string           `json:"name"`
	TradeName        string           `json:"tradeName"`
	Email            string           `json:"email"`
	RegistrationDate string           `json:"registrationDate"`
	Type             string           `json:"type"`
	LastUpdateDate   time.Time        `json:"lastUpdateDate"`
	Locked           bool             `json:"locked"`
	MobileNumber     string           `json:"mobileNumber"`
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
