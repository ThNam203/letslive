package rtmp

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"

	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/transcode/domains"
	livestreamdto "sen1or/letslive/transcode/gateway/livestream/dto"
	"sen1or/letslive/transcode/live"
	"sen1or/letslive/transcode/response"
)

func TestMain(m *testing.M) {
	logger.Init(logger.Error)
	os.Exit(m.Run())
}

type statusUpdate struct {
	vodId       string
	status      domains.VODStatus
	playbackURL string
	duration    *int64
}

// fakeEnder records calls in order; events is shared with fakeSession.
type fakeEnder struct {
	events  *[]string
	vodId   *uuid.UUID
	endErr  *response.Response[any]
	ended   livestreamdto.EndLivestreamRequestDTO
	updates []statusUpdate
	// updateFailures is how many UpdateVODStatus calls fail before one succeeds.
	updateFailures int
}

func (f *fakeEnder) EndLivestream(_ context.Context, _ string, d livestreamdto.EndLivestreamRequestDTO) (*uuid.UUID, *response.Response[any]) {
	*f.events = append(*f.events, "end")
	f.ended = d
	return f.vodId, f.endErr
}

func (f *fakeEnder) UpdateVODStatus(_ context.Context, vodId string, status domains.VODStatus, playbackUrl, _ string, duration *int64) error {
	*f.events = append(*f.events, "update")
	f.updates = append(f.updates, statusUpdate{vodId: vodId, status: status, playbackURL: playbackUrl, duration: duration})
	if len(f.updates) <= f.updateFailures {
		return errors.New("vod service unavailable")
	}
	return nil
}

type fakeSession struct {
	events   *[]string
	duration time.Duration
	closeErr error
}

func (f *fakeSession) Duration() time.Duration { return f.duration }

func (f *fakeSession) Close(context.Context) error {
	*f.events = append(*f.events, "close")
	return f.closeErr
}

const testPrefix = "https://api.example.test/transcode/vods"

func TestFinishStream(t *testing.T) {
	vodId := uuid.Must(uuid.NewV4())
	tests := []struct {
		name       string
		vodId      *uuid.UUID
		endErr     *response.Response[any]
		closeErr   error
		wantEvents []string
		wantStatus domains.VODStatus // "" when no update is expected
	}{
		{name: "recorded", vodId: &vodId, wantEvents: []string{"end", "close", "update"}, wantStatus: domains.VODStatusReady},
		{name: "recording failed", vodId: &vodId, closeErr: errors.New("no segment uploaded"), wantEvents: []string{"end", "close", "update"}, wantStatus: domains.VODStatusFailed},
		{name: "never started", vodId: &vodId, closeErr: live.ErrNothingRecorded, wantEvents: []string{"end", "close", "update"}, wantStatus: domains.VODStatusFailed},
		{name: "end failed", endErr: &response.Response[any]{Message: "boom"}, wantEvents: []string{"end", "close"}},
		{name: "no vod id", wantEvents: []string{"end", "close"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var events []string
			ender := &fakeEnder{events: &events, vodId: tt.vodId, endErr: tt.endErr}
			sess := &fakeSession{events: &events, duration: 61500 * time.Millisecond, closeErr: tt.closeErr}

			finishStream(context.Background(), ender, sess, "ls-1", testPrefix)

			if len(events) != len(tt.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tt.wantEvents)
			}
			for i := range events {
				if events[i] != tt.wantEvents[i] {
					t.Fatalf("events = %v, want %v", events, tt.wantEvents)
				}
			}
			if ender.ended.VODStatus != "processing" || ender.ended.PlaybackURL != nil || ender.ended.Duration != 62 {
				t.Errorf("end request = %+v, want vodStatus processing, no playback URL, duration 62", ender.ended)
			}
			if ender.ended.EndedAt.IsZero() {
				t.Error("end request has no endedAt")
			}

			if tt.wantStatus == "" {
				if len(ender.updates) != 0 {
					t.Errorf("updates = %+v, want none", ender.updates)
				}
				return
			}
			u := ender.updates[0]
			if u.vodId != vodId.String() || u.status != tt.wantStatus {
				t.Errorf("update = %+v, want vod %s status %s", u, vodId, tt.wantStatus)
			}
			if tt.wantStatus == domains.VODStatusReady {
				if u.playbackURL != testPrefix+"/ls-1/index.m3u8" || u.duration == nil || *u.duration != 62 {
					t.Errorf("ready update = %+v, want playback URL and duration 62", u)
				}
			} else if u.playbackURL != "" {
				t.Errorf("failed update has a playback URL %q", u.playbackURL)
			}
		})
	}
}

func noRetryDelay(t *testing.T) {
	t.Helper()
	old := vodStatusRetryDelays
	vodStatusRetryDelays = []time.Duration{0, 0}
	t.Cleanup(func() { vodStatusRetryDelays = old })
}

// A failed status update is retried, so a brief vod service outage does not
// leave the VOD stuck in processing.
func TestFinishStreamRetriesVODStatus(t *testing.T) {
	noRetryDelay(t)
	vodId := uuid.Must(uuid.NewV4())
	tests := []struct {
		failures    int
		wantUpdates int
	}{
		{failures: 1, wantUpdates: 2},
		{failures: 2, wantUpdates: 3},
		{failures: 5, wantUpdates: 3}, // gives up after the last retry
	}
	for _, tt := range tests {
		var events []string
		ender := &fakeEnder{events: &events, vodId: &vodId, updateFailures: tt.failures}
		finishStream(context.Background(), ender, &fakeSession{events: &events, duration: time.Second}, "ls-1", testPrefix)

		if len(ender.updates) != tt.wantUpdates {
			t.Errorf("%d failure(s): updates = %d, want %d", tt.failures, len(ender.updates), tt.wantUpdates)
		}
		for _, u := range ender.updates {
			if u.status != domains.VODStatusReady {
				t.Errorf("retry changed the status to %s", u.status)
			}
		}
	}
}
