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
	ErrInvalidInput   = NewServerError(http.StatusBadRequest, "Input invalid.")
	ErrInvalidPayload = NewServerError(http.StatusBadRequest, "Payload invalid.")
	ErrInvalidPath    = NewServerError(http.StatusBadRequest, "Invalid path.")

	ErrUnauthorized = NewServerError(http.StatusUnauthorized, "Unauthorized.")

	ErrForbidden = NewServerError(http.StatusForbidden, "Forbidden.")

	ErrUserNotFound  = NewServerError(http.StatusNotFound, "User not found.")
	ErrRouteNotFound = NewServerError(http.StatusNotFound, "Requested endpoint not found.")

	ErrImageTooLarge = NewServerError(http.StatusRequestEntityTooLarge, "Image exceeds 10mb limit.")

	ErrDatabaseQuery  = NewServerError(http.StatusInternalServerError, "Error querying database, please try again.")
	ErrDatabaseIssue  = NewServerError(http.StatusInternalServerError, "Database issue, please try again.")
	ErrInternalServer = NewServerError(http.StatusInternalServerError, "Something went wrong.")
)
