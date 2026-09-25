package domains

import "errors"

// Domain errors returned by repositories, services, gateways and utils.
//
// Nothing below the handler decides an HTTP status, a business code or an
// i18n key: those belong to the transport layer, which maps these values in
// serviceresponse.FromError. Wrap them with fmt.Errorf("%w") to add context;
// the mapping uses errors.Is.
var (
	ErrInvalidInput  = errors.New("invalid input")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrDatabaseQuery = errors.New("database query failed")
	ErrDatabaseIssue = errors.New("database issue")
	ErrInternal      = errors.New("internal error")

	ErrAuthNotFound             = errors.New("auth record not found")
	ErrAuthAlreadyExists        = errors.New("auth record already exists")
	ErrEmailOrPasswordIncorrect = errors.New("email or password incorrect")
	ErrPasswordNotMatch         = errors.New("password does not match")
	ErrCaptchaFailed            = errors.New("captcha verification failed")
	ErrRefreshTokenNotFound     = errors.New("refresh token not found")
	ErrSignUpOTPNotFound        = errors.New("sign up otp not found")
	ErrSignUpOTPExpired         = errors.New("sign up otp expired")
	ErrSignUpOTPAlreadyUsed     = errors.New("sign up otp already used")
	ErrFailedToCreateSignUpOTP  = errors.New("failed to create sign up otp")
	ErrFailedToSendVerification = errors.New("failed to send verification")
	ErrAccountDisabled          = errors.New("account disabled")
)

// DownstreamError carries a failure reported by another service verbatim.
//
// The auth sign-up flow calls the user service, whose errors (a taken
// username, for one) must reach the client with the downstream code and i18n
// key intact. The fields are plain values rather than a response envelope so
// that domains stays free of the transport package.
type DownstreamError struct {
	StatusCode int
	Code       int
	Key        string
	Message    string
}

func (e *DownstreamError) Error() string { return e.Message }
