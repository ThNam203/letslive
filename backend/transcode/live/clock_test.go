package live

import (
	"testing"
	"time"
)

func TestClockVideo90k(t *testing.T) {
	c := NewClock(time.Unix(0, 0), time.Second)
	if got := c.Video90k(1500 * time.Millisecond); got != 45000 {
		t.Errorf("Video90k(1.5s) = %d, want 45000", got)
	}
}

func TestClockAudio(t *testing.T) {
	c := NewClock(time.Unix(0, 0), time.Second)
	if got := c.Audio(2*time.Second, 44100); got != 44100 {
		t.Errorf("Audio(2s, 44100) = %d, want 44100", got)
	}
}

func TestClockNTP(t *testing.T) {
	start := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	c := NewClock(start, 0)
	if got, want := c.NTP(90000), start.Add(time.Second); !got.Equal(want) {
		t.Errorf("NTP(90000) = %v, want %v", got, want)
	}
}

func TestClockLongStreamNoOverflow(t *testing.T) {
	c := NewClock(time.Unix(0, 0), 0)
	day := 24 * time.Hour
	if got, want := c.Video90k(day), int64(24*3600*90000); got != want {
		t.Errorf("Video90k(24h) = %d, want %d", got, want)
	}
	if got, want := c.Audio(day, 48000), int64(24*3600*48000); got != want {
		t.Errorf("Audio(24h, 48000) = %d, want %d", got, want)
	}
	start := time.Unix(1_000_000, 0)
	c = NewClock(start, 0)
	if got, want := c.NTP(int64(24*3600*90000)), start.Add(day); !got.Equal(want) {
		t.Errorf("NTP(24h) = %v, want %v", got, want)
	}
}

func TestTSUnwrapperWrap(t *testing.T) {
	c := NewClock(time.Unix(0, 0), 0)
	u := c.NewTSUnwrapper()
	first := u.Rel(0x1FFFFFF00)
	second := u.Rel(0x100)
	if got := second - first; got != 0x200 {
		t.Errorf("second - first = %#x, want 0x200", got)
	}
}

func TestTSUnwrapperSharesClockZero(t *testing.T) {
	// zero = 1s = 90000 ticks; first raw ts at 3s (270000) maps to 2s on the clock.
	c := NewClock(time.Unix(0, 0), time.Second)
	u := c.NewTSUnwrapper()
	if got := u.Rel(270000); got != 180000 {
		t.Errorf("Rel(first) = %d, want 180000", got)
	}
	if got := u.Rel(270000 + 3000); got != 183000 {
		t.Errorf("Rel(second) = %d, want 183000", got)
	}
}
