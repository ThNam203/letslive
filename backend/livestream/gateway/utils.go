package gateway

import (
	"context"
	"errors"
	"net/http"
	"sen1or/letslive/livestream/pkg/logger"

	"github.com/gofrs/uuid/v5"
)

// SetRequestIDHeader extracts the request ID from the context and adds it to the request header.
func SetRequestIDHeader(ctx context.Context, req *http.Request) error {
	if req == nil {
		return errors.New("no request found")
	}

	if ctx == nil {
		return errors.New("no context found")
	}

	v, ok := ctx.Value("requestId").(string)
	if !ok || len(v) == 0 {
		logger.Warnf(ctx, "no requestId found when setting request id header, creating new one")
		newId, err := uuid.NewGen().NewV4()
		if err != nil {
			newId = uuid.Nil
		}
		v = newId.String()
	}

	req.Header.Set("X-Request-ID", v)
	return nil
}
