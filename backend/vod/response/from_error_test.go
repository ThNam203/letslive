package response

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"sen1or/letslive/vod/domains"

	"github.com/go-playground/validator/v10"
)

// Every domain error must produce exactly the envelope the pre-refactor code
// built at the failure site. A mismatch here is a client-visible API change.
func TestFromErrorMapsToOriginalTemplate(t *testing.T) {
	cases := []struct {
		err      error
		expected ResponseTemplate
	}{
		{domains.ErrInvalidInput, RES_ERR_INVALID_INPUT},
		{domains.ErrInvalidPayload, RES_ERR_INVALID_PAYLOAD},
		{domains.ErrForbidden, RES_ERR_FORBIDDEN},
		{domains.ErrDatabaseQuery, RES_ERR_DATABASE_QUERY},
		{domains.ErrDatabaseIssue, RES_ERR_DATABASE_ISSUE},
		{domains.ErrInternal, RES_ERR_INTERNAL_SERVER},
		{domains.ErrVODNotFound, RES_ERR_VOD_NOT_FOUND},
		{domains.ErrVODCreateFailed, RES_ERR_VOD_CREATE_FAILED},
		{domains.ErrVODUpdateFailed, RES_ERR_VOD_UPDATE_FAILED},
		{domains.ErrVODViewThreshold, RES_ERR_VOD_VIEW_THRESHOLD},
		{domains.ErrCommentNotFound, RES_ERR_VOD_COMMENT_NOT_FOUND},
		{domains.ErrCommentCreateFailed, RES_ERR_VOD_COMMENT_CREATE_FAILED},
		{domains.ErrCommentDeleteFailed, RES_ERR_VOD_COMMENT_DELETE_FAILED},
		{domains.ErrCommentUpdateFailed, RES_ERR_VOD_COMMENT_UPDATE_FAILED},
		{domains.ErrCommentAlreadyLiked, RES_ERR_VOD_COMMENT_ALREADY_LIKED},
		{domains.ErrCommentNotLiked, RES_ERR_VOD_COMMENT_NOT_LIKED},
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

// Wrapping with %w must not change the mapping: repositories and services add
// context to errors as they propagate.
func TestFromErrorSeesThroughWrapping(t *testing.T) {
	wrapped := fmt.Errorf("getvodbyid: %w", domains.ErrVODNotFound)

	got := FromError(wrapped)
	if got.Key != RES_ERR_VOD_NOT_FOUND_KEY {
		t.Errorf("key = %q, want %q", got.Key, RES_ERR_VOD_NOT_FOUND_KEY)
	}
	if got.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", got.StatusCode, http.StatusNotFound)
	}
}

func TestFromErrorNil(t *testing.T) {
	if got := FromError(nil); got != nil {
		t.Errorf("FromError(nil) = %+v, want nil", got)
	}
}

// An error that is not a known domain error must not leak its text: it becomes
// a generic internal error.
func TestFromErrorUnknownIsInternal(t *testing.T) {
	got := FromError(errors.New("pq: connection string is not a valid dsn"))

	if got.Key != RES_ERR_INTERNAL_SERVER_KEY {
		t.Errorf("key = %q, want %q", got.Key, RES_ERR_INTERNAL_SERVER_KEY)
	}
	if got.Message != RES_ERR_INTERNAL_SERVER.Message {
		t.Errorf("message = %q, want the generic internal message", got.Message)
	}
}

// Validation failures must still reach the client with per-field details.
func TestFromErrorKeepsValidationDetails(t *testing.T) {
	type payload struct {
		Content string `validate:"required"`
	}
	validationErr := validator.New().Struct(&payload{})
	if validationErr == nil {
		t.Fatal("expected the fixture to fail validation")
	}

	got := FromError(fmt.Errorf("%w: %w", domains.ErrInvalidInput, validationErr))

	if got.Key != RES_ERR_INVALID_INPUT_KEY {
		t.Errorf("key = %q, want %q", got.Key, RES_ERR_INVALID_INPUT_KEY)
	}
	if got.ErrorDetails == nil || len(*got.ErrorDetails) != 1 {
		t.Fatalf("errorDetails = %v, want 1 entry", got.ErrorDetails)
	}
	if field := (*got.ErrorDetails)[0]["Field"]; field != "Content" {
		t.Errorf("field = %v, want Content", field)
	}
}
