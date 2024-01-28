package validation

import (
	"github.com/go-playground/validator/v10"
	"github.com/jaganathanb/dapps-api/common"
)

func IndianMobileNumberValidator(fld validator.FieldLevel) bool {

	value, ok := fld.Field().Interface().(string)
	if !ok {
		return false
	}

	return common.IndianMobileNumberValidate(value)
}
