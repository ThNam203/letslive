package live

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMediaScheduleTakesFirstKeyframeAfter5s(t *testing.T) {
	s := &mediaSchedule{min: 5 * time.Second}
	for _, c := range []struct {
		media time.Duration
		want  bool
	}{{0, false}, {4 * time.Second, false}, {6 * time.Second, true}, {70 * time.Second, false}} {
		if got := s.Want(c.media); got != c.want {
			t.Errorf("Want(%v) = %v, want %v", c.media, got, c.want)
		}
	}
	if s.Fallback() {
		t.Error("Fallback after a thumbnail was taken = true, want false")
	}
}

func TestMediaScheduleFallsBackWhenNothingTaken(t *testing.T) {
	s := &mediaSchedule{min: 5 * time.Second}
	s.Want(time.Second)
	if !s.Fallback() {
		t.Error("Fallback with nothing taken = false, want true")
	}
}

func TestWallScheduleNeverFallsBack(t *testing.T) {
	if newWallSchedule(time.Hour).Fallback() {
		t.Error("wall schedule falls back, want never")
	}
}

// A video shorter than the start delay still gets a thumbnail: its first
// keyframe.
func TestThumbnailerShortVideoUsesFirstKeyframe(t *testing.T) {
	requireFFmpeg(t)
	f := loadStreamFixture(t, "src_640x360.flv")
	st := newRecordingStorage(t)
	th := newThumbnailer(context.Background(), "upload short", "short", "ffmpeg", fixtureSource(f), st,
		&mediaSchedule{min: 5 * time.Second})

	th.Offer(keyframeAUs(t, f, 1)[0], time.Second)
	if n := len(st.Calls()); n != 0 {
		t.Fatalf("thumbnail taken before Close: %d call(s)", n)
	}
	th.Close()

	calls := st.Calls()
	if len(calls) != 1 || calls[0].Kind != "thumbnail" {
		t.Fatalf("calls = %+v, want one thumbnail", calls)
	}
	if got := th.URL(); got != calls[0].URL || got == "" {
		t.Errorf("URL = %q, want %q", got, calls[0].URL)
	}
}

func TestThumbnailerURLEmptyWhenUploadFails(t *testing.T) {
	requireFFmpeg(t)
	f := loadStreamFixture(t, "src_640x360.flv")
	st := newRecordingStorage(t)
	st.before = func(storageCall) error { return errors.New("minio down") }
	th := NewThumbnailer(context.Background(), "livestream fail", "fail", "ffmpeg", fixtureSource(f), st, time.Hour)

	th.Offer(keyframeAUs(t, f, 1)[0], 0)
	th.Close()
	if got := th.URL(); got != "" {
		t.Errorf("URL after a failed upload = %q, want empty", got)
	}
}
