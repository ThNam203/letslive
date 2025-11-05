package serviceresponse

const (
	RES_SUCC_SENT_VERIFICATION_EMAIL_CODE = 10000
	RES_SUCC_EMAIL_VERIFIED_CODE          = 10001
	RES_SUCC_LOGIN_CODE                   = 10002
	RES_SUCC_SIGN_UP_CODE                 = 10003
)

const (
	RES_SUCC_SENT_VERIFICATION_EMAIL_KEY = "res_succ_sent_verification_email"
	RES_SUCC_EMAIL_VERIFIED_KEY          = "res_succ_email_verified"
	RES_SUCC_LOGIN_KEY                   = "res_succ_login"
	RES_SUCC_SIGN_UP_KEY                 = "res_succ_sign_up"
)

var (
	RES_SUCC_SENT_VERIFICATION = ResponseTemplate{
		Success:    true,
		StatusCode: 201,
		Code:       RES_SUCC_SENT_VERIFICATION_EMAIL_CODE,
		Key:        RES_SUCC_SENT_VERIFICATION_EMAIL_KEY,
		Message:    "A verification has been sent to verify your email.",
	}

	RES_SUCC_EMAIL_VERIFIED = ResponseTemplate{
		Success:    true,
		StatusCode: 200,
		Code:       RES_SUCC_EMAIL_VERIFIED_CODE,
		Key:        RES_SUCC_EMAIL_VERIFIED_KEY,
		Message:    "Your email had been verified successfully, please continue to sign up.",
	}

	RES_SUCC_LOGIN = ResponseTemplate{
		Success:    true,
		StatusCode: 200,
		Code:       RES_SUCC_LOGIN_CODE,
		Key:        RES_SUCC_LOGIN_KEY,
		Message:    "Login successfully!",
	}

	RES_SUCC_SIGN_UP = ResponseTemplate{
		Success:    true,
		StatusCode: 201,
		Code:       RES_SUCC_LOGIN_CODE,
		Key:        RES_SUCC_LOGIN_KEY,
		Message:    "Sign up successfully!",
	}
)
