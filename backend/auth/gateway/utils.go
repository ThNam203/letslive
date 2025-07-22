package gateway

import (
	"context"
	"errors"
	"net/http"
)

type contextKey string

const requestIDKey = contextKey("requestID")

// SetRequestIDHeader extracts the request ID from the context and adds it to the request header.
func SetRequestIDHeader(ctx context.Context, req *http.Request) error {
	if req == nil {
		return errors.New("no request found")
	}

	if ctx == nil {
		return errors.New("no context found")
	}

	v, ok := ctx.Value(requestIDKey).(string)
	if !ok || len(v) == 0 || v == "" {
		return errors.New("no requestId/correlationId found")
	}

	req.Header.Set("X-Request-ID", v)
	return nil
}
