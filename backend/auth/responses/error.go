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
	ErrInvalidInput      = NewServiceErrorResponse(http.StatusBadRequest, "Input invalid.")
	ErrInvalidPayload    = NewServiceErrorResponse(http.StatusBadRequest, "Payload invalid.")
	ErrAuthAlreadyExists = NewServiceErrorResponse(http.StatusBadRequest, "Email has already been registered.")
	ErrCaptchaFailed     = NewServiceErrorResponse(http.StatusBadRequest, "Failed to verify CAPTCHA, please try again.")
	ErrPasswordNotMatch  = NewServiceErrorResponse(http.StatusBadRequest, "Old password does not match.")

	ErrUnauthorized             = NewServiceErrorResponse(http.StatusUnauthorized, "Unauthorized.")
	ErrSignUpOTPExpired         = NewServiceErrorResponse(http.StatusUnauthorized, "OTP code has expired, please issue a new one.")
	ErrEmailOrPasswordIncorrect = NewServiceErrorResponse(http.StatusUnauthorized, "Username or password incorrect.")

	ErrForbidden = NewServiceErrorResponse(http.StatusForbidden, "Forbidden.")

	ErrAuthNotFound         = NewServiceErrorResponse(http.StatusNotFound, "Authentication credentials not found.")
	ErrRefreshTokenNotFound = NewServiceErrorResponse(http.StatusNotFound, "Refresh token not found.")
	ErrSignUpOTPNotFound    = NewServiceErrorResponse(http.StatusNotFound, "OTP code not found.")
	ErrRouteNotFound        = NewServiceErrorResponse(http.StatusNotFound, "Requested endpoint not found.")

	ErrSignUpOTPAlreadyUsed    = NewServiceErrorResponse(http.StatusConflict, "The OTP has already been used.")
	ErrFailedToCreateSignUpOTP = NewServiceErrorResponse(http.StatusConflict, "Failed to generate the OTP, please try again later.")

	ErrDatabaseQuery            = NewServiceErrorResponse(http.StatusInternalServerError, "Something went wrong.")
	ErrDatabaseIssue            = NewServiceErrorResponse(http.StatusInternalServerError, "Something went wrong.")
	ErrInternalServer           = NewServiceErrorResponse(http.StatusInternalServerError, "Something went wrong.")
	ErrFailedToSendVerification = NewServiceErrorResponse(http.StatusInternalServerError, "Failed to send email verification, please try again later.")
)
