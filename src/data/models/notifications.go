package models

type Notifications struct {
	BaseModel
	Message     string `json:"message"`
	MessageType string `json:"messageType"`
	Title       string `json:"title"`
	IsRead      bool   `json:"isRead"`
}
