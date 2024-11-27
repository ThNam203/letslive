package controllers

import (
	"errors"
	"fmt"
	"os"
	"sen1or/lets-live/auth/domains"
	"sen1or/lets-live/auth/repositories"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
)

type TokenControllerConfig struct {
	RefreshTokenMaxAge int
	AccessTokenMaxAge  int
}

type MyCustomClaims struct {
	UserId string `json:"userId"`
	jwt.RegisteredClaims
}

type TokenController struct {
	repo   repositories.RefreshTokenRepository
	config TokenControllerConfig
}

func NewTokenController(repo repositories.RefreshTokenRepository, cfg TokenControllerConfig) *TokenController {
	return &TokenController{
		repo:   repo,
		config: cfg,
	}
}

// generate the refresh token with access token (for login and signup)
func (c *TokenController) GenerateTokenPair(userId string) (
	*struct {
		RefreshToken       string
		RefreshTokenMaxAge int
		AccessToken        string
		AccessTokenMaxAge  int
	}, error) {

	refreshToken, err := c.generateRefreshToken(userId)
	if err != nil {
		return nil, err
	}

	accessToken, err := c.generateAccessToken(userId)
	if err != nil {
		return nil, err
	}

	return &struct {
		RefreshToken       string
		RefreshTokenMaxAge int
		AccessToken        string
		AccessTokenMaxAge  int
	}{
		RefreshToken:       refreshToken,
		RefreshTokenMaxAge: c.config.RefreshTokenMaxAge,
		AccessToken:        accessToken,
		AccessTokenMaxAge:  c.config.AccessTokenMaxAge,
	}, nil
}

// create a new access token for the refresh token
// the process is called "refresh token"
func (c *TokenController) RefreshToken(refreshToken string) (
	*struct {
		AccessToken       string
		AccessTokenMaxAge int
	}, error) {
	myClaims := MyCustomClaims{}
	parsedToken, err := jwt.NewParser().ParseWithClaims(refreshToken, &myClaims, func(t *jwt.Token) (interface{}, error) {
		return os.Getenv("REFRESH_TOKEN_SECRET"), nil
	})
	if err != nil && !parsedToken.Valid {
		return nil, errors.New("token parsing failed")
	}

	accessToken, err := c.generateAccessToken(myClaims.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %s", err)
	}

	return &struct {
		AccessToken       string
		AccessTokenMaxAge int
	}{
		AccessToken:       accessToken,
		AccessTokenMaxAge: c.config.AccessTokenMaxAge,
	}, nil
}

func (c *TokenController) generateRefreshToken(userId string) (string, error) {
	refreshTokenExpiresDuration := time.Duration(c.config.RefreshTokenMaxAge) * time.Second
	refreshTokenExpiresAt := time.Now().Add(refreshTokenExpiresDuration)
	myClaims := MyCustomClaims{
		userId,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshTokenExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "letslive",
			Subject:   "auth",
		},
	}
	unsignedRefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, myClaims)

	refreshToken, err := unsignedRefreshToken.SignedString([]byte(os.Getenv("REFRESH_TOKEN_SECRET")))
	if err != nil {
		return "", err
	}

	userIdUUID := uuid.FromStringOrNil(userId)
	refreshTokenRecord := &domains.RefreshToken{
		UserID:    userIdUUID,
		Value:     refreshToken,
		ExpiresAt: refreshTokenExpiresAt,
	}

	if err := c.repo.Create(refreshTokenRecord); err != nil {
		return "", err
	}

	return refreshToken, nil
}

func (c *TokenController) generateAccessToken(userId string) (string, error) {
	accessTokenDuration := time.Duration(c.config.AccessTokenMaxAge) * time.Second
	accessTokenExpiresAt := time.Now().Add(accessTokenDuration)
	myClaims := MyCustomClaims{
		userId,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessTokenExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "letslive",
			Subject:   "auth",
		},
	}
	unsignedAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, myClaims)

	accessToken, err := unsignedAccessToken.SignedString([]byte(os.Getenv("ACCESS_TOKEN_SECRET")))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func (c *TokenController) RevokeTokenByValue(tokenValue string) error {
	token, err := c.repo.FindByValue(tokenValue)
	if err != nil {
		return err
	}

	now := time.Now()
	token.RevokedAt = &now

	err = c.repo.Update(token)
	return err
}

func (c *TokenController) RevokeAllTokensOfUser(userID uuid.UUID) error {
	return c.repo.RevokeAllTokensOfUser(userID)
}
