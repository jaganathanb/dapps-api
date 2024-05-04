package models

type Notifications struct {
	BaseModel
	Message string `json:"message"`
	Type    string `json:"type"`
	Title   string `json:"title"`
	IsRead  bool   `json:"isRead"`
}
