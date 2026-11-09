package live

import (
	"time"

	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts"
)

const videoClockRate = 90000

// Clock gives every rendition one time base. Zero is the DTS of the first
// source keyframe; start is the wall-clock time that instant maps to.
type Clock struct {
	start time.Time
	zero  time.Duration
}

// NewClock returns a Clock whose zero is the RTMP timestamp zero and whose
// wall-clock origin is start.
func NewClock(start time.Time, zero time.Duration) *Clock {
	return &Clock{start: start, zero: zero}
}

// ticks converts d to a count of 1/rate second units. The whole seconds and
// the sub-second remainder are scaled separately so long streams cannot
// overflow int64.
func ticks(d time.Duration, rate int64) int64 {
	secs := int64(d / time.Second)
	rem := int64(d % time.Second)
	return secs*rate + rem*rate/int64(time.Second)
}

// Video90k converts an RTMP timestamp to the 90 kHz video timeline.
func (c *Clock) Video90k(t time.Duration) int64 {
	return ticks(t-c.zero, videoClockRate)
}

// Audio converts an RTMP timestamp to an audio timeline of sampleRate Hz.
func (c *Clock) Audio(t time.Duration, sampleRate int) int64 {
	return ticks(t-c.zero, int64(sampleRate))
}

// NTP converts a 90 kHz timestamp to wall-clock time.
func (c *Clock) NTP(pts90k int64) time.Time {
	secs := pts90k / videoClockRate
	rem := pts90k % videoClockRate
	offset := time.Duration(secs)*time.Second + time.Duration(rem)*time.Second/videoClockRate
	return c.start.Add(offset)
}

// TSUnwrapper turns the 33-bit wrapping MPEG-TS timestamps of one ffmpeg
// output into a monotonic 90 kHz timeline shared with the Clock. It is not
// safe for concurrent use.
type TSUnwrapper struct {
	zero90k int64
	decoder mpegts.TimeDecoder
	started bool
	offset  int64
}

// NewTSUnwrapper returns an unwrapper whose output is on c's timeline.
func (c *Clock) NewTSUnwrapper() *TSUnwrapper {
	u := &TSUnwrapper{zero90k: ticks(c.zero, videoClockRate)}
	u.decoder.Initialize()
	return u
}

// Rel returns raw relative to the Clock's zero, with wraparound removed.
func (u *TSUnwrapper) Rel(raw int64) int64 {
	if !u.started {
		u.started = true
		u.offset = raw - u.zero90k
	}
	return u.decoder.Decode(raw) + u.offset
}
