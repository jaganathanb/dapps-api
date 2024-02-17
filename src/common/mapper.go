package common

import (
	"encoding/json"

	"github.com/jaganathanb/dapps-api/data/models"
)

func TypeConverter[T any](data any) (*T, error) {
	var result T

	dataJson, err := json.Marshal(&data)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(dataJson, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func prepareGstDTO(gst models.Gst) {
	panic("unimplemented")
}
