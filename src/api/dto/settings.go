package dto

type SettingsPayload struct {
	Id          int    `json:"id"`
	Crontab     string `json:"crontab"`
	GstUsername string `json:"gstUsername"`
	GstPassword string `json:"gstPassword"`
	GstBaseUrl  string `json:"gstBaseUrl"`
}
