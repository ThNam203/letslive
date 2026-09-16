package response

import (
	"errors"
	"fmt"
	"testing"

	"sen1or/letslive/livestream/domains"
)

// Every domain error must produce exactly the envelope the pre-refactor code
// built at the failure site. A mismatch here is a client-visible API change.
func TestFromErrorMapsToOriginalTemplate(t *testing.T) {
	cases := []struct {
		err      error
		expected ResponseTemplate
	}{
		{domains.ErrInvalidPayload, RES_ERR_INVALID_PAYLOAD},
		{domains.ErrForbidden, RES_ERR_FORBIDDEN},
		{domains.ErrDatabaseQuery, RES_ERR_DATABASE_QUERY},
		{domains.ErrDatabaseIssue, RES_ERR_DATABASE_ISSUE},
		{domains.ErrInternal, RES_ERR_INTERNAL_SERVER},
		{domains.ErrLivestreamNotFound, RES_ERR_LIVESTREAM_NOT_FOUND},
		{domains.ErrLivestreamCreateFailed, RES_ERR_LIVESTREAM_CREATE_FAILED},
		{domains.ErrLivestreamUpdateFailed, RES_ERR_LIVESTREAM_UPDATE_FAILED},
		{domains.ErrLivestreamUpdateAfterEnded, RES_ERR_LIVESTREAM_UPDATE_AFTER_ENDED},
		{domains.ErrEndAlreadyEndedLivestream, RES_ERR_END_ALREADY_ENDED_LIVESTREAM},
		{domains.ErrVODCreateFailed, RES_ERR_VOD_CREATE_FAILED},
	}

	for _, tc := range cases {
		t.Run(tc.err.Error(), func(t *testing.T) {
			got := FromError(tc.err)
			if got.StatusCode != tc.expected.StatusCode {
				t.Errorf("status = %d, want %d", got.StatusCode, tc.expected.StatusCode)
			}
			if got.Code != tc.expected.Code {
				t.Errorf("code = %d, want %d", got.Code, tc.expected.Code)
			}
			if got.Key != tc.expected.Key {
				t.Errorf("key = %q, want %q", got.Key, tc.expected.Key)
			}
			if got.Success {
				t.Error("success = true, want false")
			}
		})
	}
}

func TestFromErrorSeesThroughWrapping(t *testing.T) {
	got := FromError(fmt.Errorf("getbyid: %w", domains.ErrLivestreamNotFound))
	if got.Key != RES_ERR_LIVESTREAM_NOT_FOUND_KEY {
		t.Errorf("key = %q, want %q", got.Key, RES_ERR_LIVESTREAM_NOT_FOUND_KEY)
	}
}

func TestFromErrorNil(t *testing.T) {
	if got := FromError(nil); got != nil {
		t.Errorf("FromError(nil) = %+v, want nil", got)
	}
}

// An unrecognised error must not leak its text to the client.
func TestFromErrorUnknownIsInternal(t *testing.T) {
	got := FromError(errors.New("pq: connection string is not a valid dsn"))
	if got.Key != RES_ERR_INTERNAL_SERVER_KEY {
		t.Errorf("key = %q, want %q", got.Key, RES_ERR_INTERNAL_SERVER_KEY)
	}
	if got.Message != RES_ERR_INTERNAL_SERVER.Message {
		t.Errorf("message = %q, want the generic internal message", got.Message)
	}
}
