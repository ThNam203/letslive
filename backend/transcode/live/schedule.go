package live

import "time"

// ThumbSchedule decides which offered keyframes become thumbnails. Want is
// asked only while the thumbnailer is idle; true means the keyframe is taken.
type ThumbSchedule interface {
	Want(media time.Duration) bool
	// Fallback reports whether, at Close, the first keyframe should be used
	// because nothing was taken.
	Fallback() bool
}

// wallSchedule follows the wall clock, as a livestream unfolds: the first
// thumbnail after thumbnailStartDelay, then one every `every`.
type wallSchedule struct {
	notBefore time.Time
	every     time.Duration
	last      time.Time
	taken     bool
}

func newWallSchedule(every time.Duration) *wallSchedule {
	return &wallSchedule{notBefore: time.Now().Add(thumbnailStartDelay), every: every}
}

func (s *wallSchedule) Want(time.Duration) bool {
	now := time.Now()
	if now.Before(s.notBefore) || (s.taken && now.Sub(s.last) < s.every) {
		return false
	}
	s.taken, s.last = true, now
	return true
}

func (s *wallSchedule) Fallback() bool { return false }

// mediaSchedule follows the video's own time, as a file is read faster than
// real time: one thumbnail, at the first keyframe at or after min.
type mediaSchedule struct {
	min   time.Duration
	taken bool
}

func (s *mediaSchedule) Want(media time.Duration) bool {
	if s.taken || media < s.min {
		return false
	}
	s.taken = true
	return true
}

func (s *mediaSchedule) Fallback() bool { return !s.taken }
