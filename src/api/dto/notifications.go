package dto

import (
	"time"

	"github.com/jaganathanb/dapps-api/constants"
)

type NotificationsPayload struct {
	Id          int                               `json:"id"`
	Message     string                            `json:"message"`
	MessageType constants.NotificationMessageType `json:"messageType"`
	Title       string                            `json:"title"`
	IsRead      bool                              `json:"isRead"`
	DeletedAt   *time.Time                        `json:"deleted_at"`
}
