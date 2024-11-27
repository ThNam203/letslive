package controllers

import (
	"sen1or/lets-live/auth/domains"
	"sen1or/lets-live/auth/repositories"
	"time"

	"github.com/gofrs/uuid/v5"
)

type VerifyTokenController struct {
	repo repositories.VerifyTokenRepository
}

func NewVerifyTokenController(repo repositories.VerifyTokenRepository) *VerifyTokenController {
	return &VerifyTokenController{
		repo: repo,
	}
}

func (c *VerifyTokenController) Create(userID uuid.UUID) (*domains.VerifyToken, error) {
	token, _ := uuid.NewGen().NewV4()
	newToken := &domains.VerifyToken{
		Token:     token.String(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
		UserID:    userID,
	}

	err := c.repo.Create(*newToken)
	if err != nil {
		return nil, err
	}

	return newToken, nil
}

func (c *VerifyTokenController) GetByValue(token string) (*domains.VerifyToken, error) {
	record, err := c.repo.GetByValue(token)
	if err != nil {
		return nil, err
	}

	return record, nil
}

//func (c *VerifyTokenController) Update(verifyToken *domains.VerifyToken) error {
//	err := c.repo.(verifyToken)
//	return err
//}

func (c *VerifyTokenController) DeleteByID(tokenID uuid.UUID) error {
	return c.repo.DeleteByID(tokenID)
}

func (c *VerifyTokenController) DeleteByValue(tokenID uuid.UUID) error {
	return c.repo.DeleteByID(tokenID)
}
