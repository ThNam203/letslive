package services

import (
	"context"
	"os"
	"sen1or/letslive/auth/config"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/auth/pkg/logger"
	serviceresponse "sen1or/letslive/auth/responses"
	"sen1or/letslive/auth/types"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	repo   domains.RefreshTokenRepository
	config config.JWT
}

func NewJWTService(repo domains.RefreshTokenRepository, cfg config.JWT) *JWTService {
	return &JWTService{
		repo:   repo,
		config: cfg,
	}
}

// generate the refresh token with access token (for login and signup)
func (c *JWTService) GenerateTokenPair(ctx context.Context, userId string) (*types.TokenPairInformation, *serviceresponse.ServiceErrorResponse) {
	refreshToken, err := c.generateRefreshToken(ctx, userId)
	if err != nil {
		return nil, serviceresponse.ErrInternalServer
	}

	accessToken, err := c.generateAccessToken(userId)
	if err != nil {
		return nil, serviceresponse.ErrInternalServer
	}

	return &types.TokenPairInformation{
		RefreshToken:       refreshToken,
		RefreshTokenMaxAge: c.config.RefreshTokenMaxAge,
		AccessToken:        accessToken,
		AccessTokenMaxAge:  c.config.AccessTokenMaxAge,
	}, nil
}

// create a new access token for the refresh token
// the process is called "refresh token"
func (c *JWTService) RefreshToken(refreshToken string) (*types.AccessTokenInformation, *serviceresponse.ServiceErrorResponse) {
	myClaims := types.MyClaims{}
	parsedToken, err := jwt.NewParser().ParseWithClaims(refreshToken, &myClaims, func(t *jwt.Token) (any, error) {
		return []byte(os.Getenv("REFRESH_TOKEN_SECRET")), nil
	})

	if err != nil {
		logger.Errorf("token parsing failed: %s", err)
		return nil, serviceresponse.ErrUnauthorized
	} else if !parsedToken.Valid {
		logger.Errorf("token not valid")
		return nil, serviceresponse.ErrUnauthorized
	}

	accessToken, err := c.generateAccessToken(myClaims.UserId)
	if err != nil {
		logger.Errorf("failed to refresh token: %s", err)
		return nil, serviceresponse.ErrUnauthorized
	}

	return &types.AccessTokenInformation{
		AccessToken:       accessToken,
		AccessTokenMaxAge: c.config.AccessTokenMaxAge,
	}, nil
}

func (c *JWTService) generateRefreshToken(ctx context.Context, userId string) (string, error) {
	refreshTokenExpiresDuration := time.Duration(c.config.RefreshTokenMaxAge) * time.Second
	refreshTokenExpiresAt := time.Now().Add(refreshTokenExpiresDuration)
	myClaims := types.MyClaims{
		UserId:   userId,
		Consumer: c.config.Consumer,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshTokenExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    c.config.Issuer,
			Subject:   c.config.Subject,
		},
	}
	unsignedRefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, myClaims)

	refreshToken, err := unsignedRefreshToken.SignedString([]byte(os.Getenv("REFRESH_TOKEN_SECRET")))
	if err != nil {
		return "", err
	}

	userIdUUID := uuid.FromStringOrNil(userId)
	refreshTokenRecord := &domains.RefreshToken{
		UserId:    userIdUUID,
		Token:     refreshToken,
		ExpiresAt: refreshTokenExpiresAt,
	}

	if err := c.repo.Insert(ctx, refreshTokenRecord); err != nil {
		return "", err
	}

	return refreshToken, nil
}

func (c *JWTService) generateAccessToken(userId string) (string, error) {
	accessTokenDuration := time.Duration(c.config.AccessTokenMaxAge) * time.Second
	accessTokenExpiresAt := time.Now().Add(accessTokenDuration)
	myClaims := types.MyClaims{
		UserId:   userId,
		Consumer: c.config.Consumer,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessTokenExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    c.config.Issuer,
			Subject:   c.config.Subject,
		},
	}
	unsignedAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, myClaims)

	accessToken, err := unsignedAccessToken.SignedString([]byte(os.Getenv("ACCESS_TOKEN_SECRET")))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func (c *JWTService) RevokeTokenByValue(ctx context.Context, tokenValue string) error {
	token, err := c.repo.FindByValue(ctx, tokenValue)
	if err != nil {
		return err
	}

	now := time.Now()
	token.RevokedAt = &now

	err = c.repo.Update(ctx, token)
	return err
}

func (c *JWTService) RevokeAllTokensOfUser(ctx context.Context, userID uuid.UUID) error {
	return c.repo.RevokeAllTokensOfUser(ctx, userID)
}
