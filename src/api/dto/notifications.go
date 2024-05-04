package dto

import "time"

type NotificationsPayload struct {
	Id        int        `json:"id"`
	Message   string     `json:"message"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	IsRead    bool       `json:"isRead"`
	DeletedAt *time.Time `json:"deleted_at"`
}
