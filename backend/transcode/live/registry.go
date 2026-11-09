package live

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gofrs/uuid/v5"
)

// Registry maps livestream ids to the Streams being served.
type Registry struct {
	mu      sync.RWMutex
	streams map[string]*Stream
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{streams: make(map[string]*Stream)}
}

// Add registers s under its id, which must be a UUID.
func (r *Registry) Add(s *Stream) error {
	if s == nil {
		return errors.New("nil stream")
	}
	key, err := canonicalID(s.ID())
	if err != nil {
		return fmt.Errorf("livestream id %q: %w", s.ID(), err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, dup := r.streams[key]; dup {
		return fmt.Errorf("livestream %s is already registered", key)
	}
	r.streams[key] = s
	return nil
}

// Remove unregisters id; later requests for it get 404. It does not close the
// stream. Unknown or malformed ids are ignored, since none can be registered.
func (r *Registry) Remove(id string) {
	key, err := canonicalID(id)
	if err != nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.streams, key)
}

// Handler serves "/{id}/index.m3u8" (master playlist) and
// "/{id}/{rendition}/{file}" (that rendition's muxer, query preserved).
// Other paths get 404, a non-UUID id 400, and methods other than GET 405.
func (r *Registry) Handler() http.Handler {
	return http.HandlerFunc(r.serveHTTP)
}

func (r *Registry) serveHTTP(w http.ResponseWriter, req *http.Request) {
	if !allowGET(w, req) {
		return
	}

	parts := strings.Split(strings.TrimPrefix(req.URL.Path, "/"), "/")
	if len(parts) < 2 || len(parts) > 3 {
		http.NotFound(w, req)
		return
	}
	key, err := canonicalID(parts[0])
	if err != nil {
		http.Error(w, "invalid livestream id", http.StatusBadRequest)
		return
	}

	r.mu.RLock()
	s := r.streams[key]
	r.mu.RUnlock()
	if s == nil {
		http.NotFound(w, req)
		return
	}
	s.serve(w, req, parts[1:])
}

func canonicalID(id string) (string, error) {
	u, err := uuid.FromString(id)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
