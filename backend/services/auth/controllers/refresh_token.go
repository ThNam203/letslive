package controllers

import (
	"os"
	"sen1or/lets-live/auth/config"
	"sen1or/lets-live/auth/domains"
	"sen1or/lets-live/auth/repositories"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
)

type RefreshTokenController struct {
	repo repositories.RefreshTokenRepository
}

func NewRefreshTokenController(repo repositories.RefreshTokenRepository) *RefreshTokenController {
	return &RefreshTokenController{
		repo: repo,
	}
}

func (c *RefreshTokenController) GenerateTokenPair(userId uuid.UUID) (
	*struct {
		RefreshToken          string
		AccessToken           string
		RefreshTokenExpiresAt time.Time
		AccessTokenExpiresAt  time.Time
	}, error) {
	refreshTokenExpiresDuration, err := time.ParseDuration(config.MyConfig.Tokens.RefreshTokenExpiresDuration)
	if err != nil {
		return nil, err
	}
	accessTokenExpiresDuration, err := time.ParseDuration(config.MyConfig.Tokens.AccessTokenExpiresDuration)
	if err != nil {
		return nil, err
	}

	refreshTokenExpiresAt := time.Now().Add(refreshTokenExpiresDuration)
	accessTokenExpiresAt := time.Now().Add(accessTokenExpiresDuration)

	unsignedRefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":    userId.String(),
		"expiresAt": refreshTokenExpiresAt,
	})

	unsignedAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":    userId.String(),
		"expiresAt": accessTokenExpiresAt,
	})

	refreshToken, err := unsignedRefreshToken.SignedString([]byte(os.Getenv("REFRESH_TOKEN_SECRET")))
	accessToken, err := unsignedAccessToken.SignedString([]byte(os.Getenv("ACCESS_TOKEN_SECRET")))

	if err != nil {
		return nil, err
	}

	refreshTokenRecord, err := c.createRefreshTokenObject(refreshToken, refreshTokenExpiresAt, userId)

	if err != nil {
		return nil, err
	}

	if err := c.repo.Create(refreshTokenRecord); err != nil {
		return nil, err
	}

	return &struct {
		RefreshToken          string
		AccessToken           string
		RefreshTokenExpiresAt time.Time
		AccessTokenExpiresAt  time.Time
	}{
		RefreshToken:          refreshToken,
		AccessToken:           accessToken,
		RefreshTokenExpiresAt: refreshTokenExpiresAt,
		AccessTokenExpiresAt:  accessTokenExpiresAt,
	}, nil
}

func (c *RefreshTokenController) createRefreshTokenObject(signedRefreshToken string, expiresAt time.Time, userId uuid.UUID) (*domains.RefreshToken, error) {
	uuid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	refreshToken := &domains.RefreshToken{
		ID:        uuid,
		UserID:    userId,
		Value:     signedRefreshToken,
		ExpiresAt: expiresAt,
	}

	return refreshToken, nil
}

func (c *RefreshTokenController) RevokeTokenByValue(tokenValue string) error {
	token, err := c.repo.FindByValue(tokenValue)
	if err != nil {
		return err
	}

	now := time.Now()
	token.RevokedAt = &now

	err = c.repo.Update(token)
	return err
}

func (c *RefreshTokenController) RevokeAllTokensOfUser(userID uuid.UUID) error {
	return c.repo.RevokeAllTokensOfUser(userID)
}
