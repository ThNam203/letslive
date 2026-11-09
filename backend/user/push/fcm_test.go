package push

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"sen1or/letslive/shared/pkg/logger"

	"firebase.google.com/go/v4/messaging"
)

func TestMain(m *testing.M) {
	logger.Init(logger.Debug)
	os.Exit(m.Run())
}

type fakeMulticaster struct {
	got  []*messaging.MulticastMessage
	resp func(tokens []string) *messaging.BatchResponse
	err  error
}

func (f *fakeMulticaster) SendEachForMulticast(_ context.Context, m *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
	f.got = append(f.got, m)
	if f.err != nil {
		return nil, f.err
	}
	return f.resp(m.Tokens), nil
}

func allSuccess(tokens []string) *messaging.BatchResponse {
	resp := &messaging.BatchResponse{SuccessCount: len(tokens)}
	for range tokens {
		resp.Responses = append(resp.Responses, &messaging.SendResponse{Success: true})
	}
	return resp
}

func TestBuildMulticastHasAlertSoundAndDataForBothPlatforms(t *testing.T) {
	msg := Message{Title: "T", Body: "B", Data: map[string]string{"type": "gift_received"}}

	m := buildMulticast([]string{"a"}, msg)

	if m.Data["type"] != "gift_received" {
		t.Errorf("data = %v", m.Data)
	}
	if m.APNS == nil || m.APNS.Payload.Aps.Alert.Title != "T" || m.APNS.Payload.Aps.Alert.Body != "B" || m.APNS.Payload.Aps.Sound == "" {
		t.Errorf("apns config incomplete: %+v", m.APNS)
	}
	if m.Android == nil || m.Android.Priority != "high" || m.Android.Notification == nil {
		t.Errorf("android config incomplete: %+v", m.Android)
	}
}

func TestGatewayChunksTokensAndMapsResults(t *testing.T) {
	tokens := make([]string, maxTokensPerBatch+1)
	for i := range tokens {
		tokens[i] = fmt.Sprintf("tok-%d", i)
	}
	fm := &fakeMulticaster{resp: allSuccess}
	gw := &FCMGateway{client: fm}

	results, err := gw.SendToTokens(context.Background(), tokens, Message{})

	if err != nil {
		t.Fatal(err)
	}
	if len(fm.got) != 2 || len(fm.got[0].Tokens) != maxTokensPerBatch || len(fm.got[1].Tokens) != 1 {
		t.Fatalf("batches = %d", len(fm.got))
	}
	if len(results) != len(tokens) || results[0].Status != StatusSent {
		t.Fatalf("results = %d, first %v", len(results), results[0].Status)
	}
}

func TestGatewayMarksUnclassifiedErrorsAsFailed(t *testing.T) {
	fm := &fakeMulticaster{resp: func(tokens []string) *messaging.BatchResponse {
		return &messaging.BatchResponse{FailureCount: 1, Responses: []*messaging.SendResponse{{Error: errors.New("boom")}}}
	}}
	gw := &FCMGateway{client: fm}

	results, err := gw.SendToTokens(context.Background(), []string{"a"}, Message{})

	if err != nil || results[0].Status != StatusFailed {
		t.Fatalf("results = %+v, err = %v", results, err)
	}
}

func TestGatewayReturnsTransportError(t *testing.T) {
	gw := &FCMGateway{client: &fakeMulticaster{err: errors.New("down")}}
	if _, err := gw.SendToTokens(context.Background(), []string{"a"}, Message{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestNewFromEnvDisabledWhenCredentialUnset(t *testing.T) {
	t.Setenv(EnvServiceAccountJSON, "")

	sender, closeFn, err := NewFromEnv(context.Background(), &fakeTokenStore{}, nil)
	defer closeFn()

	if err != nil {
		t.Fatal(err)
	}
	if _, ok := sender.(Disabled); !ok {
		t.Fatalf("sender = %T, want Disabled", sender)
	}
}

func TestNewFromEnvFailsOnMalformedCredential(t *testing.T) {
	for name, raw := range map[string]string{
		"not json":       "{nope",
		"wrong type":     `{"type":"authorized_user"}`,
		"missing fields": `{"type":"service_account","project_id":"p"}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv(EnvServiceAccountJSON, raw)

			_, _, err := NewFromEnv(context.Background(), &fakeTokenStore{}, nil)

			if err == nil || !strings.Contains(err.Error(), EnvServiceAccountJSON) {
				t.Fatalf("err = %v, want a clear error naming %s", err, EnvServiceAccountJSON)
			}
		})
	}
}
