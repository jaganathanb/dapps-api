package services

import (
	"sync"

	"github.com/jaganathanb/dapps-api/api/dto"
	"github.com/jaganathanb/dapps-api/common"
	"github.com/jaganathanb/dapps-api/config"
	"github.com/jaganathanb/dapps-api/data/db"
	"github.com/jaganathanb/dapps-api/data/models"
	"github.com/jaganathanb/dapps-api/pkg/logging"
	"gorm.io/gorm"
)

type NotificationsService struct {
	logger          logging.Logger
	cfg             *config.Config
	database        *gorm.DB
	streamerService *StreamerService
}

var notificationsService *NotificationsService
var notificationsServiceOnce sync.Once

func NewNotificationsService(cfg *config.Config) *NotificationsService {
	notificationsServiceOnce.Do(func() {
		database := db.GetDb()
		logger := logging.NewLogger(cfg)
		notificationsService = &NotificationsService{
			cfg:             cfg,
			database:        database,
			logger:          logger,
			streamerService: NewStreamerService(cfg),
		}
	})

	return notificationsService
}

// Add notifications
func (s *NotificationsService) AddNotification(req *dto.NotificationsPayload) (bool, error) {
	tx := s.database.Begin()

	notifications := models.Notifications{
		Message:     req.Message,
		MessageType: string(req.MessageType),
		Title:       req.Title,
		UserId:      req.UserId,
		IsRead:      req.IsRead,
		BaseModel: models.BaseModel{
			CreatedBy: req.BaseDto.CreatedBy,
		},
	}

	err := tx.Model(&models.Notifications{}).Where("id = ?", req.Id).Save(&notifications).Error

	if err != nil {
		tx.Rollback()
		return false, nil
	}

	tx.Commit()

	s.streamerService.StreamData(StreamMessage{Code: "NOTIFICATION", Message: req.Message, MessageType: req.MessageType, Title: req.Title})

	return true, nil
}

// Get notifications
func (s *NotificationsService) GetNotifications(userId int) ([]dto.NotificationsPayload, error) {
	notifications := []models.Notifications{}

	err := s.database.Model(&models.Notifications{}).Where("deleted_at is null").Find(&notifications).Error

	if err != nil {
		return []dto.NotificationsPayload{}, nil
	}

	result := []dto.NotificationsPayload{}
	for _, notif := range notifications {
		res, _ := common.TypeConverter[dto.NotificationsPayload](notif)

		result = append(result, *res)
	}

	return result, nil
}

// Update notifications
func (s *NotificationsService) UpdateNotifications(req *dto.NotificationsPayload) (bool, error) {
	var notifications models.Notifications
	err := s.database.Model(&models.Notifications{}).Where("id = ?", req.Id).First(&notifications).Error

	if err != nil {
		return false, err
	}

	tx := s.database.Begin()

	notifications.IsRead = req.IsRead
	notifications.DeletedAt = req.DeletedAt

	err = tx.Model(&models.Notifications{}).Where("id = ?", req.Id).Updates(notifications).Error

	if err != nil {
		tx.Rollback()
		return false, nil
	}

	tx.Commit()

	return true, nil
}

// Delete notifications
func (s *NotificationsService) DeleteNotifications(req *dto.NotificationsPayload) (bool, error) {
	var notifications models.Notifications
	err := s.database.Model(&models.Notifications{}).Where("id = ?", req.Id).First(&notifications).Error

	if err != nil {
		return false, err
	}

	tx := s.database.Begin()

	err = tx.Model(&models.Notifications{}).Where("id = ?", req.Id).Delete(&notifications).Error

	if err != nil {
		tx.Rollback()
		return false, nil
	}

	tx.Commit()

	return true, nil
}
