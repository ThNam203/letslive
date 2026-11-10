package live

import (
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
)

// mustLiveMode is LiveMode for tests that know their registry is valid.
func mustLiveMode(t *testing.T, reg *Registry) Mode {
	t.Helper()
	m, err := LiveMode(reg, testLiveConfig())
	if err != nil {
		t.Fatalf("LiveMode: %v", err)
	}
	return m
}

func TestNoOutputAcceptsEverything(t *testing.T) {
	out := noOutput{names: []string{"source", "360p"}}
	if err := out.WriteVideo(0, time.Now(), 0, [][]byte{{0x65}}); err != nil {
		t.Errorf("WriteVideo = %v, want nil", err)
	}
	if err := out.WriteAudio(time.Now(), 0, [][]byte{{0x01}}); err != nil {
		t.Errorf("WriteAudio = %v, want nil", err)
	}
	for r := 0; r < 4; r++ {
		if !out.Ready(r) {
			t.Errorf("Ready(%d) = false, want true", r)
		}
	}
	out.DropRendition(1)
	out.Unpublish()
	out.Close()
	if name, ok := out.Rendition(1); !ok || name != "360p" {
		t.Errorf("Rendition(1) = %q, %v, want 360p, true", name, ok)
	}
	if _, ok := out.Rendition(2); ok {
		t.Error("Rendition(2) ok, want out of range")
	}
}

func TestLiveModeNeedsRegistry(t *testing.T) {
	if _, err := LiveMode(nil, testLiveConfig()); err == nil {
		t.Fatal("LiveMode(nil) = nil error, want one")
	}
}

func TestLiveModeOpenRegistersAndUnpublishRemoves(t *testing.T) {
	f := loadStreamFixture(t, "src_1280x720.flv")
	reg := NewRegistry()
	id := uuid.Must(uuid.NewV4()).String()

	m := mustLiveMode(t, reg)
	out, err := m.open(id, []Rendition{{Name: "source"}}, f.audio)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(out.Close)

	if !registered(reg, id) {
		t.Fatal("open did not register the stream")
	}
	out.Unpublish()
	if registered(reg, id) {
		t.Fatal("Unpublish left the stream registered")
	}
}

func registered(reg *Registry, id string) bool {
	key, err := canonicalID(id)
	if err != nil {
		return false
	}
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	return reg.streams[key] != nil
}

func TestLiveModeOpenDuplicateID(t *testing.T) {
	f := loadStreamFixture(t, "src_1280x720.flv")
	m := mustLiveMode(t, NewRegistry())
	id := uuid.Must(uuid.NewV4()).String()
	out, err := m.open(id, []Rendition{{Name: "source"}}, f.audio)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(out.Close)
	if _, err := m.open(id, []Rendition{{Name: "source"}}, f.audio); err == nil {
		t.Fatal("second open with the same id = nil error, want one")
	}
}

func TestLiveModeLabel(t *testing.T) {
	if got := mustLiveMode(t, NewRegistry()).label("abc"); got != "livestream abc" {
		t.Errorf("label = %q, want %q", got, "livestream abc")
	}
}
