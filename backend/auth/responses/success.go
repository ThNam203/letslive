package serviceresponse

import (
	"net/http"
)

type ServerSuccessResponse struct {
	StatusCode int
	Message    string
	Data       any
}

func NewServerSuccessResponse(statusCode int, message string, data any) *ServerSuccessResponse {
	return &ServerSuccessResponse{
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	}
}

var (
	SuccessSentVerification = NewServerSuccessResponse(http.StatusCreated, "A verification has been sent to verify your email.", nil)
	SuccessEmailVerified    = NewServerSuccessResponse(http.StatusOK, "Your email had been verified successfully, please continue to sign up.", nil)
)
