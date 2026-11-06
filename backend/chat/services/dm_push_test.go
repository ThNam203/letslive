package services

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"sen1or/letslive/chat/domains"
	"sen1or/letslive/shared/pkg/logger"
)

type pushCall struct {
	recipients     []string
	senderName     string
	text           string
	conversationID string
}

// fakePushClient records calls and returns err.
type fakePushClient struct {
	calls []pushCall
	err   error
}

func (f *fakePushClient) SendDmPush(_ context.Context, recipientIDs []string, senderName, text, conversationID string) error {
	f.calls = append(f.calls, pushCall{recipientIDs, senderName, text, conversationID})
	return f.err
}

func TestSendDmPushesToAllParticipantsExceptSender(t *testing.T) {
	push := &fakePushClient{}
	s := &DmMessageService{push: push}

	s.pushDm(context.Background(), []string{"a", "b", "c"}, "a", "alice",
		&domains.DmMessage{Type: domains.DmMessageTypeText, Text: "hi"}, "conv1")

	if len(push.calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(push.calls))
	}
	got := push.calls[0]
	if !slices.Equal(got.recipients, []string{"b", "c"}) || got.senderName != "alice" ||
		got.text != "hi" || got.conversationID != "conv1" {
		t.Fatalf("call = %+v", got)
	}
}

func TestSendDmPushSkippedWhenSenderIsAlone(t *testing.T) {
	push := &fakePushClient{}
	s := &DmMessageService{push: push}

	s.pushDm(context.Background(), []string{"a"}, "a", "alice", &domains.DmMessage{Text: "hi"}, "conv1")

	if len(push.calls) != 0 {
		t.Fatalf("calls = %d, want 0", len(push.calls))
	}
}

func TestPushFailureDoesNotPropagate(t *testing.T) {
	logger.Init(logger.Debug) // pushDm logs the failure
	push := &fakePushClient{err: errors.New("user service down")}
	s := &DmMessageService{push: push}

	// pushDm has no return value: a failure can only be logged.
	s.pushDm(context.Background(), []string{"a", "b"}, "a", "alice", &domains.DmMessage{Text: "hi"}, "conv1")

	if len(push.calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(push.calls))
	}
}

func TestImageOnlyMessageUsesFixedBody(t *testing.T) {
	for _, text := range []string{"Sent an image", "Sent 3 images"} {
		msg := &domains.DmMessage{Type: domains.DmMessageTypeImage, Text: text, ImageURLs: []string{"u"}}
		if got := dmPushBody(msg); got != ImagePushBody {
			t.Errorf("dmPushBody(%q) = %q, want %q", text, got, ImagePushBody)
		}
	}
	captioned := &domains.DmMessage{Type: domains.DmMessageTypeImage, Text: "look at this", ImageURLs: []string{"u"}}
	if got := dmPushBody(captioned); got != "look at this" {
		t.Errorf("captioned image body = %q", got)
	}
}

func TestLongTextIsTruncatedTo500Chars(t *testing.T) {
	msg := &domains.DmMessage{Type: domains.DmMessageTypeText, Text: strings.Repeat("é", 800)}

	got := dmPushBody(msg)

	if n := len([]rune(got)); n != MaxPushBodyLen {
		t.Fatalf("length = %d, want %d", n, MaxPushBodyLen)
	}
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("body does not end with an ellipsis")
	}
	short := &domains.DmMessage{Type: domains.DmMessageTypeText, Text: strings.Repeat("a", 500)}
	if dmPushBody(short) != short.Text {
		t.Fatal("a 500 char text must not be truncated")
	}
}
