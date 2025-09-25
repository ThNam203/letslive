package services

import (
	"context"
	"os"
	"sen1or/letslive/auth/config"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/auth/pkg/logger"
	serviceresponse "sen1or/letslive/auth/response"
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
func (c *JWTService) GenerateTokenPair(ctx context.Context, userId string) (*types.TokenPairInformation, *serviceresponse.Response[any]) {
	refreshToken, err := c.generateRefreshToken(ctx, userId)
	if err != nil {
		return nil, err
	}

	accessToken, err := c.generateAccessToken(userId)
	if err != nil {
		return nil, err
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
func (c *JWTService) RefreshToken(ctx context.Context, refreshToken string) (*types.AccessTokenInformation, *serviceresponse.Response[any]) {
	myClaims := types.MyClaims{}
	parsedToken, err := jwt.NewParser().ParseWithClaims(refreshToken, &myClaims, func(t *jwt.Token) (any, error) {
		return []byte(os.Getenv("REFRESH_TOKEN_SECRET")), nil
	})

	if err != nil {
		logger.Errorf(ctx, "token parsing failed: %s", err)
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	} else if !parsedToken.Valid {
		logger.Errorf(ctx, "token not valid")
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		)
	}

	accessToken, genErr := c.generateAccessToken(myClaims.UserId)
	if genErr != nil {
		logger.Errorf(ctx, "failed to refresh token: %s", genErr)
		return nil, genErr
	}

	return &types.AccessTokenInformation{
		AccessToken:       accessToken,
		AccessTokenMaxAge: c.config.AccessTokenMaxAge,
	}, nil
}

func (c *JWTService) generateRefreshToken(ctx context.Context, userId string) (string, *serviceresponse.Response[any]) {
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
		return "", serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
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

func (c *JWTService) generateAccessToken(userId string) (string, *serviceresponse.Response[any]) {
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
		return "", serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	return accessToken, nil
}

func (c *JWTService) RevokeTokenByValue(ctx context.Context, tokenValue string) *serviceresponse.Response[any] {
	token, err := c.repo.FindByValue(ctx, tokenValue)
	if err != nil {
		return err
	}

	now := time.Now()
	token.RevokedAt = &now

	err = c.repo.Update(ctx, token)
	return err
}

func (c *JWTService) RevokeAllTokensOfUser(ctx context.Context, userID uuid.UUID) *serviceresponse.Response[any] {
	return c.repo.RevokeAllTokensOfUser(ctx, userID)
}
