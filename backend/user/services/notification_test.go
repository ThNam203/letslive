package services

import (
	"context"
	"errors"
	"testing"

	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/shared/pkg/realtime"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/dto"

	"github.com/gofrs/uuid/v5"
)

func init() { logger.Init(logger.Debug) }

type fakeNotificationRepo struct {
	domains.NotificationRepository
	createErr error
}

func (f fakeNotificationRepo) Create(_ context.Context, n domains.Notification) (*domains.Notification, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	n.Id = uuid.Must(uuid.NewV4())
	return &n, nil
}

type fakePublisher struct{}

func (fakePublisher) Publish(context.Context, realtime.Topic, string, any) error { return nil }

type recordingPush struct {
	calls []domains.Notification
	err   error
	panic bool
}

func (r *recordingPush) Send(_ context.Context, _ uuid.UUID, n domains.Notification) error {
	if r.panic {
		panic("push exploded")
	}
	r.calls = append(r.calls, n)
	return r.err
}

func createRequest() dto.CreateNotificationRequestDTO {
	return dto.CreateNotificationRequestDTO{
		UserId: uuid.Must(uuid.NewV4()).String(), Type: "gift_received", Title: "t", Message: "m",
	}
}

func TestCreateNotificationSendsPushForStoredRow(t *testing.T) {
	push := &recordingPush{}
	svc := NewNotificationService(fakeNotificationRepo{}, fakePublisher{}, push)

	created, err := svc.CreateNotification(context.Background(), createRequest())

	if err != nil {
		t.Fatal(err)
	}
	if len(push.calls) != 1 || push.calls[0].Id != created.Id {
		t.Fatalf("push calls = %+v, want one for notification %s", push.calls, created.Id)
	}
}

func TestSendFailureDoesNotFailNotificationCreation(t *testing.T) {
	push := &recordingPush{err: errors.New("fcm down")}
	svc := NewNotificationService(fakeNotificationRepo{}, fakePublisher{}, push)

	created, err := svc.CreateNotification(context.Background(), createRequest())

	if err != nil || created == nil {
		t.Fatalf("created = %v, err = %v; want success", created, err)
	}
}

func TestPanickingPushSenderDoesNotFailNotificationCreation(t *testing.T) {
	svc := NewNotificationService(fakeNotificationRepo{}, fakePublisher{}, &recordingPush{panic: true})

	created, err := svc.CreateNotification(context.Background(), createRequest())

	if err != nil || created == nil {
		t.Fatalf("created = %v, err = %v; want success", created, err)
	}
}

func TestNoPushWhenRowWasNotCreated(t *testing.T) {
	push := &recordingPush{}
	svc := NewNotificationService(fakeNotificationRepo{createErr: errors.New("db")}, fakePublisher{}, push)

	if _, err := svc.CreateNotification(context.Background(), createRequest()); err == nil {
		t.Fatal("expected the repository error")
	}
	if len(push.calls) != 0 {
		t.Fatal("push sent for a notification that was not stored")
	}
}
