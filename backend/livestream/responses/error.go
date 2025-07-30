package serviceresponse

import (
	"net/http"
)

type ServiceErrorResponse struct {
	StatusCode int
	Message    string
}

func (e *ServiceErrorResponse) Error() string {
	return e.Message
}

func NewServiceErrorResponse(statusCode int, message string) *ServiceErrorResponse {
	return &ServiceErrorResponse{StatusCode: statusCode, Message: message}
}

var (
	ErrInvalidInput               = NewServiceErrorResponse(http.StatusBadRequest, "Input invalid.")
	ErrInvalidPayload             = NewServiceErrorResponse(http.StatusBadRequest, "Payload invalid.")
	ErrInvalidPath                = NewServiceErrorResponse(http.StatusBadRequest, "Invalid path.")
	ErrLivestreamUpdateAfterEnded = NewServiceErrorResponse(http.StatusBadRequest, "Failed to update, the livestream has ended.")

	ErrUnauthorized = NewServiceErrorResponse(http.StatusUnauthorized, "Unauthorized.")

	ErrForbidden = NewServiceErrorResponse(http.StatusForbidden, "Forbidden.")

	ErrLivestreamNotFound = NewServiceErrorResponse(http.StatusNotFound, "Livestream not found.")
	ErrVODNotFound        = NewServiceErrorResponse(http.StatusNotFound, "VOD not found.")
	ErrRouteNotFound      = NewServiceErrorResponse(http.StatusNotFound, "Requested endpoint not found.")

	ErrEndAnAlreadyEndedLivestream = NewServiceErrorResponse(http.StatusConflict, "The livestream has already been ended.")

	ErrImageTooLarge = NewServiceErrorResponse(http.StatusRequestEntityTooLarge, "Image exceeds 10mb limit.")

	ErrDatabaseQuery   = NewServiceErrorResponse(http.StatusInternalServerError, "Something went wrong.")
	ErrInternalServer  = NewServiceErrorResponse(http.StatusInternalServerError, "Something went wrong.")
	ErrQueryScanFailed = NewServiceErrorResponse(http.StatusInternalServerError, "Failed to scan or collect database rows.")

	ErrLivestreamCreateFailed = NewServiceErrorResponse(http.StatusInternalServerError, "Failed to create livestream record.")
	ErrLivestreamUpdateFailed = NewServiceErrorResponse(http.StatusInternalServerError, "Failed to update livestream record.")
	ErrVODCreateFailed        = NewServiceErrorResponse(http.StatusInternalServerError, "Failed to create vod record.")
	ErrVODUpdateFailed        = NewServiceErrorResponse(http.StatusInternalServerError, "Failed to update vod record.")
)
