package validation

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	Property string `json:"property"`
	Tag      string `json:"tag"`
	Value    string `json:"value"`
	Message  string `json:"message"`
}

func GetValidationErrors(err error) *[]ValidationError {
	var validationErrors []ValidationError
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, err := range err.(validator.ValidationErrors) {
			var el ValidationError
			el.Property = err.Field()
			el.Tag = err.Tag()
			el.Value = fmt.Sprintf("%v", err.Value())
			el.Message = getMessageForTag(err.Tag())
			validationErrors = append(validationErrors, el)
		}
		return &validationErrors
	}
	return nil
}

func getMessageForTag(s string) string {
	switch s {
	case "required":
		return "Field is required"
	case "gstins":
		return "Not all the GSTIN is in right format"
	case "gstin":
		return "GSTIN is not in right format"
	}

	return ""
}
