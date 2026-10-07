package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"sen1or/letslive/chat/response"
	"sen1or/letslive/chat/services"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/golang-jwt/jwt/v5"
)

type accessClaims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

// GetUserIDFromCookie reads the user id from the ACCESS_TOKEN cookie. Kong has
// already verified the signature; like the Node service, an expired token or
// one without a userId is still rejected.
func GetUserIDFromCookie(r *http.Request) (string, *response.Response[any]) {
	unauthorized := response.NewResponseFromTemplate[any](response.RES_ERR_UNAUTHORIZED, nil, nil, nil)

	cookie, err := r.Cookie("ACCESS_TOKEN")
	if err != nil || cookie.Value == "" {
		logger.Debugf(r.Context(), "missing or invalid credentials in cookie")
		return "", unauthorized
	}

	var claims accessClaims
	if _, _, err := jwt.NewParser().ParseUnverified(cookie.Value, &claims); err != nil {
		logger.Debugf(r.Context(), "invalid access token: %v", err)
		return "", unauthorized
	}
	if claims.UserID == "" {
		return "", unauthorized
	}
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return "", unauthorized
	}
	return claims.UserID, nil
}

// express.json's default body limit
const maxBodyBytes = 100 << 10

var ErrInvalidPayload = errors.New("invalid payload")

// DecodeBody decodes a JSON object body into dst. An empty body leaves dst
// untouched, as Express gives handlers an empty object then.
func DecodeBody(r *http.Request, dst any) error {
	body, err := io.ReadAll(http.MaxBytesReader(nil, r.Body, maxBodyBytes))
	if err != nil {
		return ErrInvalidPayload
	}
	if len(bytes.TrimSpace(body)) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, dst); err != nil {
		return ErrInvalidPayload
	}
	return nil
}

func InvalidPayload() *response.Response[any] {
	return response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil)
}

func InvalidInput() *response.Response[any] {
	return response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil)
}

// The Node handlers checked raw body values with JavaScript semantics
// (typeof, truthiness). Bodies are decoded into json.RawMessage fields and
// inspected with these helpers so the same values pass and fail.

// AsString returns the value when raw is a JSON string.
func AsString(raw json.RawMessage) (string, bool) {
	var s string
	if len(raw) == 0 || raw[0] != '"' {
		return "", false
	}
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", false
	}
	return s, true
}

// IsTruthy is JavaScript truthiness of a JSON value; an absent field is
// undefined and therefore falsy.
func IsTruthy(raw json.RawMessage) bool {
	var value any
	if len(raw) == 0 || json.Unmarshal(raw, &value) != nil {
		return false
	}
	switch v := value.(type) {
	case nil:
		return false
	case bool:
		return v
	case float64:
		return v != 0
	case string:
		return v != ""
	default:
		return true
	}
}

// AsStringArray returns the elements when raw is an array of strings.
func AsStringArray(raw json.RawMessage) ([]string, bool) {
	var values []any
	if len(raw) == 0 || raw[0] != '[' || json.Unmarshal(raw, &values) != nil {
		return nil, false
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		s, ok := value.(string)
		if !ok {
			return nil, false
		}
		out = append(out, s)
	}
	return out, true
}

// OptionalString reads a field that may be absent, null or a string.
func OptionalString(raw json.RawMessage) services.OptionalString {
	if len(raw) == 0 {
		return services.OptionalString{}
	}
	if string(bytes.TrimSpace(raw)) == "null" {
		return services.OptionalString{Present: true}
	}
	if s, ok := AsString(raw); ok {
		return services.OptionalString{Present: true, Value: &s}
	}
	return services.OptionalString{Present: true, Invalid: true}
}
