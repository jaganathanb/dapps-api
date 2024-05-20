package dto

type SettingsPayload struct {
	BaseDto
	Crontab     string `json:"crontab"`
	GstUsername string `json:"gstUsername"`
	GstPassword string `json:"gstPassword"`
	GstBaseUrl  string `json:"gstBaseUrl"`
}
