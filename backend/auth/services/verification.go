package services

import (
	"context"
	"net/smtp"
	"os"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/auth/pkg/logger"
	serviceresponse "sen1or/letslive/auth/response"
	"sen1or/letslive/auth/utils"
	"time"
)

type VerificationService struct {
	repo domains.SignUpOTPRepository
}

func NewVerificationService(repo domains.SignUpOTPRepository) *VerificationService {
	return &VerificationService{
		repo: repo,
	}
}

func (c *VerificationService) CreateSignUpOTP(ctx context.Context, email string) (*domains.SignUpOTP, *serviceresponse.Response[any]) {
	generatedOTP, err := utils.GenerateOTP()
	if err != nil {
		return nil, err
	}

	newToken := &domains.SignUpOTP{
		Code:      generatedOTP,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Email:     email,
	}

	err = c.repo.Insert(ctx, *newToken)
	if err != nil {
		return nil, err
	}

	return newToken, nil
}

func (s VerificationService) Verify(ctx context.Context, code, email string) *serviceresponse.Response[any] {
	otp, err := s.repo.GetOTP(ctx, code, email)
	if err != nil {
		return err
	}

	if otp.UsedAt != nil {
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_SIGN_UP_OTP_ALREADY_USED,
			nil,
			nil,
			nil,
		)
	}

	if otp.ExpiresAt.Before(time.Now()) {
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_SIGN_UP_OTP_EXPIRED,
			nil,
			nil,
			nil,
		)
	}

	if err := s.repo.UpdateUsedAt(ctx, otp.Id, time.Now()); err != nil {
		return err
	}

	return nil
}

func (c *VerificationService) CreateOTPAndSendEmailVerification(ctx context.Context, verificationGateway string, userEmail string) *serviceresponse.Response[any] {
	createdToken, err := c.CreateSignUpOTP(ctx, userEmail)
	if err != nil {
		return err
	}

	smtpServer := "smtp.gmail.com:587"
	smtpUser := "letsliveglobal@gmail.com"
	smtpPassword := os.Getenv("GMAIL_APP_PASSWORD")

	from := "letsliveglobal@gmail.com"
	to := []string{userEmail}
	subject := "Lets Live Email Verification Code"

	body := `<!DOCTYPE html>
            <html>
            <head>
                <title>` + subject + `</title>
                <style>
                    /* Basic styles for better readability */
                    body { font-family: Arial, sans-serif; line-height: 1.6; color: #333333; }
                    .container { padding: 20px; border: 1px solid #dddddd; margin: 10px; max-width: 600px; }
                    .code-display {
                        background-color: #f0f8ff; /* Light Alice Blue background */
                        border: 1px dashed #add8e6; /* Light blue dashed border */
                        padding: 15px;
                        margin: 20px 0;
                        text-align: center;
                        font-size: 24px; /* Larger font size for code */
                        font-weight: bold;
                        letter-spacing: 3px; /* Space out digits slightly */
                        color: #0056b3; /* Darker blue color for code */
                    }
                    .footer { font-size: 0.9em; color: #777777; margin-top: 15px;}
                </style>
            </head>
            <body>
                <div class="container">
                    <h2>Email Verification Required</h2>
                    <p>Hello,</p>
                    <p>Please use the following verification code to complete your action with Let's Live. This code is valid for a 5 minutes.</p>

                    <p>Your verification code is:</p>
                    <div class="code-display">
                        ` + createdToken.Code + `
                    </div>

                    <p>If you did not request this verification, please disregard this email.</p>

                    <p class="footer">Best Regards,<br>The Let's Live Global Team</p>
                </div>
            </body>
            </html>`

	msg := "From: " + from + "\n" +
		"To: " + userEmail + "\n" +
		"Subject: " + subject + "\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\n\n" + // Ensures HTML is rendered
		body

	auth := smtp.PlainAuth("", smtpUser, smtpPassword, "smtp.gmail.com")

	mErr := smtp.SendMail(smtpServer, auth, from, to, []byte(msg))
	if mErr != nil {
		logger.Errorf(ctx, "failed trying to send confirmation code email to %s: %s", userEmail, mErr.Error())
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_FAILED_TO_SEND_VERIFICATION,
			nil,
			nil,
			nil,
		)
	}

	logger.Infof(ctx, "verification code email sent successfully to %s", userEmail)
	return nil // success
}
