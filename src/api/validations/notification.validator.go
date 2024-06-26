package validation

import (
	"github.com/go-playground/validator/v10"
	"github.com/jaganathanb/dapps-api/common"
)

func NotificationMessageTypeValidator(fld validator.FieldLevel) bool {
	value, ok := fld.Field().Interface().(string)
	if !ok {
		fld.Param()
		return false
	}

	return common.CheckNotificationType(value)
}
