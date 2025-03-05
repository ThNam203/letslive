package servererrors

import (
	"net/http"
)

type ServerError struct {
	StatusCode int
	Message    string
}

func (e *ServerError) Error() string {
	return e.Message
}

func NewServerError(statusCode int, message string) *ServerError {
	return &ServerError{StatusCode: statusCode, Message: message}
}

var (
	ErrInvalidInput      = NewServerError(http.StatusBadRequest, "Input invalid.")
	ErrInvalidPayload    = NewServerError(http.StatusBadRequest, "Payload invalid.")
	ErrAuthAlreadyExists = NewServerError(http.StatusBadRequest, "Email is already registered.")
	ErrPasswordNotMatch  = NewServerError(http.StatusBadRequest, "Old password does not match.")

	ErrUnauthorized             = NewServerError(http.StatusUnauthorized, "Unauthorized.")
	ErrEmailOrPasswordIncorrect = NewServerError(http.StatusUnauthorized, "Username or password incorrect.")

	ErrForbidden = NewServerError(http.StatusForbidden, "Forbidden.")

	ErrAuthNotFound         = NewServerError(http.StatusNotFound, "Authentication credentials not found.")
	ErrRefreshTokenNotFound = NewServerError(http.StatusNotFound, "Refresh token not found.")
	ErrVerifyTokenNotFound  = NewServerError(http.StatusNotFound, "Verify token not found.")
	ErrRouteNotFound        = NewServerError(http.StatusNotFound, "Requested endpoint not found.")

	ErrDatabaseQuery  = NewServerError(http.StatusInternalServerError, "Something went wrong.")
	ErrDatabaseIssue  = NewServerError(http.StatusInternalServerError, "Something went wrong.")
	ErrInternalServer = NewServerError(http.StatusInternalServerError, "Something went wrong.")
)
