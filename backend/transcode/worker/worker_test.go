package worker

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sen1or/letslive/transcode/upload"
)

func TestForceConvert(t *testing.T) {
	str := func(s string) *string { return &s }
	for _, c := range []struct {
		prev *string
		want bool
	}{
		{nil, false},
		{str(""), false},
		{str("input: upload input: intake ffmpeg: exit status 1"), true},
		{str("failed to download raw file: timeout"), false},
	} {
		if got := forceConvert(c.prev); got != c.want {
			t.Errorf("forceConvert(%v) = %v, want %v", c.prev, got, c.want)
		}
	}
}

func TestFailureMessagePrefixesInputErrors(t *testing.T) {
	in := fmt.Errorf("%w: unusable video packet", upload.ErrInput)
	if got := failureMessage(in); !strings.HasPrefix(got, inputErrPrefix) {
		t.Errorf("failureMessage(input error) = %q, want prefix %q", got, inputErrPrefix)
	}
	if got := failureMessage(errors.New("minio down")); strings.HasPrefix(got, inputErrPrefix) {
		t.Errorf("failureMessage(other error) = %q, want no input prefix", got)
	}
	if !forceConvert(ptr(failureMessage(in))) {
		t.Error("an input failure's message does not force conversion on the retry")
	}
}

// Review Focus 5: a retry must not see the playlists of a failed attempt.
func TestPrepareAttemptRemovesOldVODDir(t *testing.T) {
	pub := t.TempDir()
	dir := filepath.Join(pub, "vods", "vod-1", "0")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "stream.m3u8"), []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := prepareAttempt(pub, "vod-1"); err != nil {
		t.Fatalf("prepareAttempt: %v", err)
	}
	if _, err := os.Stat(filepath.Join(pub, "vods", "vod-1")); !os.IsNotExist(err) {
		t.Errorf("old VOD dir still there: %v", err)
	}
	if err := prepareAttempt(pub, "vod-2"); err != nil {
		t.Errorf("prepareAttempt on a missing dir = %v, want nil", err)
	}
}

func TestPrepareAttemptRejectsUnsafeID(t *testing.T) {
	pub := t.TempDir()
	for _, id := range []string{"", ".", "..", "../x", "a/b"} {
		if err := prepareAttempt(pub, id); err == nil {
			t.Errorf("prepareAttempt(%q) = nil, want an error", id)
		}
	}
}

func ptr(s string) *string { return &s }
