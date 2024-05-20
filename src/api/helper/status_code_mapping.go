package helper

import (
	"net/http"

	service_errors "github.com/jaganathanb/dapps-api/pkg/service-errors"
)

var StatusCodeMapping = map[string]int{

	// OTP
	service_errors.OptExists:   409,
	service_errors.OtpUsed:     409,
	service_errors.OtpNotValid: 400,

	// User
	service_errors.EmailExists:      409,
	service_errors.UsernameExists:   409,
	service_errors.RecordNotFound:   404,
	service_errors.PermissionDenied: 403,

	// GST
	service_errors.GstNotFound: 404,
	service_errors.GstExists:   409,
	service_errors.GstsExists:  409,
}

func TranslateErrorToStatusCode(err error) int {
	value, ok := StatusCodeMapping[err.Error()]
	if !ok {
		return http.StatusInternalServerError
	}
	return value
}
