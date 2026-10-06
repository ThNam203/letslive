package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"sen1or/letslive/realtime/auth"
	"sen1or/letslive/realtime/client"
	"sen1or/letslive/realtime/config"
	"sen1or/letslive/shared/middlewares"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/coder/websocket"
)

const readHeaderTimeout = 10 * time.Second

type Server struct {
	cfg        *config.Config
	verifier   *auth.Verifier
	hub        client.Hub
	presence   client.Presence
	isHealthy  func() bool
	baseCtx    context.Context
	cancelBase context.CancelFunc
	httpServer *http.Server
}

func NewServer(cfg *config.Config, verifier *auth.Verifier, h client.Hub, p client.Presence, isHealthy func() bool) *Server {
	baseCtx, cancelBase := context.WithCancel(context.Background())
	s := &Server{
		cfg:        cfg,
		verifier:   verifier,
		hub:        h,
		presence:   p,
		isHealthy:  isHealthy,
		baseCtx:    baseCtx,
		cancelBase: cancelBase,
	}
	// no Read/WriteTimeout: they would put deadlines on hijacked socket connections
	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.Service.APIBindAddress, cfg.Service.APIPort),
		Handler:           s.handler(),
		ReadHeaderTimeout: readHeaderTimeout,
	}
	return s
}

func (s *Server) handler() http.Handler {
	mux := http.NewServeMux()

	var health http.Handler = http.HandlerFunc(s.health)
	health = middlewares.RequestIDMiddleware(health)
	health = middlewares.LoggingMiddleware(health)
	mux.Handle("GET /v1/health", health)

	// the logging middleware's ResponseWriter does not implement http.Hijacker,
	// which websocket.Accept needs, so the upgrade route stays unwrapped
	mux.HandleFunc("GET /realtime", s.upgrade)

	return mux
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	status, natsStatus, code := "ok", "ok", http.StatusOK
	if !s.isHealthy() {
		status, natsStatus, code = "degraded", "unavailable", http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	body := map[string]any{"status": status, "checks": map[string]string{"nats": natsStatus}}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		logger.Errorf(r.Context(), "failed to write health response: %v", err)
	}
}

func (s *Server) upgrade(w http.ResponseWriter, r *http.Request) {
	userID, _ := s.verifier.UserID(r)

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: s.cfg.WebSocket.AllowedOrigins})
	if err != nil {
		logger.Debugf(r.Context(), "websocket upgrade rejected: %v", err)
		return
	}

	// the request context must not be used after the connection is hijacked
	client.Serve(s.baseCtx, conn, userID, s.hub, s.presence)
}

func (s *Server) ListenAndServe() error {
	err := s.httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// Shutdown cancels every open socket first because http.Server.Shutdown does
// not track hijacked connections.
func (s *Server) Shutdown(ctx context.Context) error {
	s.cancelBase()
	return s.httpServer.Shutdown(ctx)
}
