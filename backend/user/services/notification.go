package services

import (
	"context"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/push"

	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/shared/pkg/realtime"

	"github.com/gofrs/uuid/v5"
)

type NotificationService struct {
	notificationRepo domains.NotificationRepository
	publisher        realtime.Publisher
	push             push.PushSender
}

func NewNotificationService(
	notificationRepo domains.NotificationRepository,
	publisher realtime.Publisher,
	pushSender push.PushSender,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		publisher:        publisher,
		push:             pushSender,
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

	created, err := s.notificationRepo.Create(ctx, notification)
	if err != nil {
		return nil, err
	}

	// the row is already stored and the client refetches on reconnect, so a
	// failed push must not fail the request
	topic := realtime.UserTopic(created.UserId.String())
	if pubErr := s.publisher.Publish(ctx, topic, realtime.EventNotificationCreated, created); pubErr != nil {
		logger.Errorf(ctx, "failed to publish notification %s to %s: %v", created.Id, topic.String(), pubErr)
	}

	s.sendPush(ctx, created)

	return created, nil
}

// sendPush is best effort: the row is stored, so neither an error nor a panic
// from the sender may fail the request.
func (s NotificationService) sendPush(ctx context.Context, n *domains.Notification) {
	// PushSender is an interface: any implementation, not only push.Async, may panic
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf(ctx, "push sender panicked for notification %s user %s: %v", n.Id, n.UserId, r)
		}
	}()
	if err := s.push.Send(ctx, n.UserId, *n); err != nil {
		logger.Errorf(ctx, "failed to queue push for notification %s user %s: %v", n.Id, n.UserId, err)
	}
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
