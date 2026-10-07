package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"sen1or/letslive/chat/domains"
)

func TestValidImageURLs(t *testing.T) {
	urls := func(n, length int) []string {
		out := make([]string, n)
		for i := range out {
			out[i] = strings.Repeat("u", length)
		}
		return out
	}

	tests := []struct {
		name string
		urls []string
		want bool
	}{
		{"no images", nil, true},
		{"10 images", urls(10, 50), true},
		{"11 images", urls(11, 50), false},
		{"2048 character url", urls(1, 2048), true},
		{"2049 character url", urls(1, 2049), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validImageURLs(tt.urls); got != tt.want {
				t.Fatalf("validImageURLs = %v, want %v", got, tt.want)
			}
		})
	}
}

// The cap is checked before any lookup, so a service without repositories or
// a user gateway is enough: reaching them would panic.
func TestCreateGroupOverCapIsRejectedBeforeLookup(t *testing.T) {
	service := NewConversationService(nil, nil, nil)

	// the owner plus 100 others is one over the cap of 100
	others := make([]string, domains.MaxGroupParticipants)
	for i := range others {
		others[i] = fmt.Sprintf("user-%d", i)
	}

	_, _, err := service.Create(context.Background(), CreateConversationInput{
		Type:           domains.ConversationTypeGroup,
		CreatorID:      "owner",
		ParticipantIDs: others,
	})
	if !errors.Is(err, domains.ErrTooManyParticipants) {
		t.Fatalf("err = %v, want ErrTooManyParticipants", err)
	}
}
