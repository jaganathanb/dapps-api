package dto

import (
	"github.com/jaganathanb/dapps-api/constants"
)

type NotificationsPayload struct {
	BaseDto
	Message     string                            `json:"message"`
	MessageType constants.NotificationMessageType `json:"messageType"`
	Title       string                            `json:"title"`
	IsRead      bool                              `json:"isRead"`
	UserId      int                               `json:"userId"`
}
