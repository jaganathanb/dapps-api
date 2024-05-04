package models

type Settings struct {
	BaseModel
	Crontab     string `json:"crontab"`
	GstUsername string `json:"gstUsername"`
	GstPassword string `json:"gstPassword"`
	GstBaseUrl  string `json:"gstBaseUrl"`
}
