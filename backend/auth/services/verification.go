package services

import (
	"context"
	"net/smtp"
	"os"
	"sen1or/letslive/auth/domains"
	servererrors "sen1or/letslive/auth/errors"
	usergateway "sen1or/letslive/auth/gateway/user"
	"sen1or/letslive/auth/pkg/logger"
	"time"

	"github.com/gofrs/uuid/v5"
)

type VerificationService struct {
	repo        domains.VerifyTokenRepository
	userGateway usergateway.UserGateway
}

func NewVerificationService(repo domains.VerifyTokenRepository, userGateway usergateway.UserGateway) *VerificationService {
	return &VerificationService{
		repo:        repo,
		userGateway: userGateway,
	}
}

func (c *VerificationService) Create(userId uuid.UUID) (*domains.VerifyToken, *servererrors.ServerError) {
	token, _ := uuid.NewGen().NewV4()
	newToken := &domains.VerifyToken{
		Token:     token.String(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
		UserID:    userId,
	}

	err := c.repo.Insert(*newToken)
	if err != nil {
		return nil, err
	}

	return newToken, nil
}

func (s VerificationService) Verify(token string) *servererrors.ServerError {
	verifyToken, err := s.repo.GetByValue(token)
	if err != nil {
		return servererrors.ErrVerifyTokenNotFound
	}

	if verifyToken.ExpiresAt.Before(time.Now()) {
		return servererrors.ErrUnauthorized
	}

	errRes := s.userGateway.UpdateUserVerified(context.Background(), verifyToken.UserID.String())
	if errRes != nil {
		return servererrors.NewServerError(errRes.StatusCode, errRes.Message)
	}

	err = s.repo.DeleteByID(verifyToken.ID)
	if err != nil {
		return err
	}

	return nil
}

func (c *VerificationService) SendConfirmEmail(token domains.VerifyToken, verificationGateway string, userEmail string) {
	smtpServer := "smtp.gmail.com:587"
	smtpUser := "letsliveglobal@gmail.com"
	smtpPassword := os.Getenv("GMAIL_APP_PASSWORD")
	verificationURL := verificationGateway + `/auth/email-verify?token=` + token.Token

	from := "letsliveglobal@gmail.com"
	to := []string{userEmail}
	subject := "Lets Live Email Confirmation"
	body := `<!DOCTYPE html>
			 <html>
			 <head>
			     <title>` + subject + `</title>
			 </head>
			 <body>
			    <p>Please confirm your email address, if you did not request any verification from Let's Live, please let us know.</p>
			    <p>The verication code has a duration of 60 minutes, if exceeds please try to issue a new one.</p>

			 	<p>Click <a href="` + verificationURL + `">here</a> to confirm your email.</p>

				<p>If the link is not clickable, please try to use the below url: ` + verificationURL + `</p>

			 	<p>Best Regards</p>
			 	<p>Let's Live Global</p>
			 </body>
			 </html>`

	msg := "From: " + from + "\n" +
		"To: " + userEmail + "\n" +
		"Subject: " + subject + "\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\n\n" +
		body

	auth := smtp.PlainAuth("", smtpUser, smtpPassword, "smtp.gmail.com")

	err := smtp.SendMail(smtpServer, auth, from, to, []byte(msg))
	if err != nil {
		logger.Errorf("failed trying to send confirmation email: %s", err.Error())
	}
}
