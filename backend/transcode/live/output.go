package live

import "time"

// RenditionWriter is the part of an Output that FeedTS writes encoded video
// to. Rendition 0 is the source; Rendition reports a rendition's name and
// whether the index is in the ladder.
type RenditionWriter interface {
	WriteVideo(r int, ntp time.Time, pts90k int64, au [][]byte) error
	Rendition(r int) (name string, ok bool)
}

// Output is where viewers get a session's renditions while it runs.
type Output interface {
	RenditionWriter
	WriteAudio(ntp time.Time, pts int64, aus [][]byte) error
	Ready(r int) bool
	DropRendition(r int)
	// Unpublish stops serving viewers; Close then releases the output.
	Unpublish()
	Close()
}

// liveOutput is a Stream served through a Registry.
type liveOutput struct {
	*Stream
	reg *Registry
}

func (o liveOutput) Unpublish() { o.reg.Remove(o.id) }

// noOutput is the Output of a session nobody watches live (an upload). It
// accepts and discards everything and is always ready, so nothing waits on
// viewers.
type noOutput struct {
	names []string
}

func newNoOutput(rs []Rendition) noOutput {
	names := make([]string, len(rs))
	for i, r := range rs {
		names[i] = r.Name
	}
	return noOutput{names: names}
}

func (noOutput) WriteVideo(int, time.Time, int64, [][]byte) error { return nil }
func (noOutput) WriteAudio(time.Time, int64, [][]byte) error      { return nil }
func (noOutput) Ready(int) bool                                   { return true }
func (noOutput) DropRendition(int)                                {}
func (noOutput) Unpublish()                                       {}
func (noOutput) Close()                                           {}

func (o noOutput) Rendition(r int) (string, bool) {
	if r < 0 || r >= len(o.names) {
		return "", false
	}
	return o.names[r], true
}
