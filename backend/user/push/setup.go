package push

import (
	"context"
	"os"
	"strings"

	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/shared/pkg/realtime"

	"github.com/nats-io/nats.go"
)

// Environment variables read by NewFromEnv.
const (
	// EnvServiceAccountJSON holds the Firebase service account key (JSON).
	// Unset means push is disabled.
	EnvServiceAccountJSON = "FIREBASE_SERVICE_ACCOUNT_JSON"
	// EnvProjectID optionally overrides the key's project_id.
	EnvProjectID = "FIREBASE_PROJECT_ID"
)

// Sender is what the rest of the service uses: the notification entry point
// plus the payload-agnostic one.
type Sender interface {
	PushSender
	MessageSender
}

// NewFromEnv builds the push pipeline. With FIREBASE_SERVICE_ACCOUNT_JSON unset
// push is disabled (logged once), sends are no-ops and no NATS subscription is
// made. A set but malformed credential returns an error so the service fails at
// boot. When enabled it subscribes to presence events on conn. The returned
// function stops the worker pool and waits for queued pushes.
func NewFromEnv(ctx context.Context, tokens TokenStore, conn *nats.Conn) (Sender, func(), error) {
	raw := strings.TrimSpace(os.Getenv(EnvServiceAccountJSON))
	if raw == "" {
		logger.Infof(ctx, "push disabled: %s is not set", EnvServiceAccountJSON)
		return NewDisabled(), func() {}, nil
	}
	gateway, err := NewFCMGateway(ctx, raw, strings.TrimSpace(os.Getenv(EnvProjectID)))
	if err != nil {
		return nil, nil, err
	}

	presence := NewPresenceTracker(conn.IsConnected)
	if _, err := presence.Subscribe(conn); err != nil {
		// without presence every user counts as offline: duplicates, never lost pushes
		logger.Errorf(ctx, "push presence subscription failed, online users will still get pushes: %v", err)
	} else {
		logger.Infof(ctx, "push presence subscribed to %s", realtime.PresenceSubject)
	}

	async := NewAsync(NewService(tokens, presence, gateway), AsyncOptions{})
	logger.Infof(ctx, "push enabled (fcm)")
	return async, async.Close, nil
}
