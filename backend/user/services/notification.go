package services

import (
	"context"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/dto"

	"github.com/gofrs/uuid/v5"
)

type NotificationService struct {
	notificationRepo domains.NotificationRepository
}

func NewNotificationService(
	notificationRepo domains.NotificationRepository,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
	}
}

func (s NotificationService) GetNotifications(ctx context.Context, userId string, page int) ([]domains.Notification, int, error) {
	userUUID, err := uuid.FromString(userId)
	if err != nil {
		return nil, 0, domains.ErrInvalidInput
	}

	pageSize := 20
	return s.notificationRepo.GetByUserId(ctx, userUUID, page, pageSize)
}

func (s NotificationService) GetUnreadCount(ctx context.Context, userId string) (int, error) {
	userUUID, err := uuid.FromString(userId)
	if err != nil {
		return 0, domains.ErrInvalidInput
	}

	return s.notificationRepo.GetUnreadCount(ctx, userUUID)
}

func (s NotificationService) CreateNotification(ctx context.Context, req dto.CreateNotificationRequestDTO) (*domains.Notification, error) {
	userUUID, err := uuid.FromString(req.UserId)
	if err != nil {
		return nil, domains.ErrInvalidInput
	}

	var referenceId *uuid.UUID
	if req.ReferenceId != nil {
		parsed, err := uuid.FromString(*req.ReferenceId)
		if err != nil {
			return nil, domains.ErrInvalidInput
		}
		referenceId = &parsed
	}

	notification := domains.Notification{
		UserId:      userUUID,
		Type:        req.Type,
		Title:       req.Title,
		Message:     req.Message,
		ActionUrl:   req.ActionUrl,
		ActionLabel: req.ActionLabel,
		ReferenceId: referenceId,
	}

	return s.notificationRepo.Create(ctx, notification)
}

func (s NotificationService) MarkAsRead(ctx context.Context, notificationId, userId string) error {
	notifUUID, err1 := uuid.FromString(notificationId)
	userUUID, err2 := uuid.FromString(userId)
	if err1 != nil || err2 != nil {
		return domains.ErrInvalidInput
	}

	return s.notificationRepo.MarkAsRead(ctx, notifUUID, userUUID)
}

func (s NotificationService) MarkAllAsRead(ctx context.Context, userId string) error {
	userUUID, err := uuid.FromString(userId)
	if err != nil {
		return domains.ErrInvalidInput
	}

	return s.notificationRepo.MarkAllAsRead(ctx, userUUID)
}

func (s NotificationService) DeleteNotification(ctx context.Context, notificationId, userId string) error {
	notifUUID, err1 := uuid.FromString(notificationId)
	userUUID, err2 := uuid.FromString(userId)
	if err1 != nil || err2 != nil {
		return domains.ErrInvalidInput
	}

	return s.notificationRepo.DeleteById(ctx, notifUUID, userUUID)
}
