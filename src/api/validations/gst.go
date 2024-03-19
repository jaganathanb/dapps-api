package validation

import (
	"github.com/go-playground/validator/v10"
	"github.com/jaganathanb/dapps-api/api/dto"
	"github.com/jaganathanb/dapps-api/common"
)

func GstinsValidator(fld validator.FieldLevel) bool {
	value, ok := fld.Field().Interface().([]dto.Gst)
	if !ok {
		fld.Param()
		return false
	}

	return common.CheckGstins(value)
}

func GstinValidator(fld validator.FieldLevel) bool {
	value, ok := fld.Field().Interface().(string)
	if !ok {
		fld.Param()
		return false
	}

	return common.CheckGstin(value)
}
