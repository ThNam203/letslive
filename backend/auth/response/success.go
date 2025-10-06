package serviceresponse

const (
	RES_SUCC_SENT_VERIFICATION_EMAIL_CODE = 10000
	RES_SUCC_EMAIL_VERIFIED_CODE          = 10001
)

const (
	RES_SUCC_SENT_VERIFICATION_EMAIL_KEY = "res_succ_sent_verification_email"
	RES_SUCC_EMAIL_VERIFIED_KEY          = "res_succ_email_verified"
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
)
