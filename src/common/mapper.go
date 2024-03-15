package common

import (
	"encoding/json"
	"strings"

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

func MergeStruc[T any](st1 T, st2 T) *T {
	conf := new(T) // New config
	*conf = st1    // Initialize with defaults

	st2Str, _ := json.Marshal(st2)
	err := json.NewDecoder(strings.NewReader(string(st2Str))).Decode(&conf)
	if err != nil {
		panic(err)
	}

	return conf
}
