package repository

import (
	"os"
	"sen1or/lets-live/server/config"
	"sen1or/lets-live/server/domain"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type postgresRefreshTokenRepo struct {
	db gorm.DB
}

func NewRefreshTokenRepository(conn gorm.DB) domain.RefreshTokenRepository {
	return &postgresRefreshTokenRepo{
		db: conn,
	}
}

func (r *postgresRefreshTokenRepo) RevokeByValue(tokenValue string) error {
	var refreshToken domain.RefreshToken
	result := r.db.First(&refreshToken, "value = ?", tokenValue)

	if result.Error != nil {
		return result.Error
	}

	refreshToken.Revoked = false
	result = r.db.Save(refreshToken)

	return result.Error
}

func (r *postgresRefreshTokenRepo) Create(tokenRecord domain.RefreshToken) error {
	result := r.db.Create(&tokenRecord)
	return result.Error
}

func (r *postgresRefreshTokenRepo) FindByValue(tokenValue string) (*domain.RefreshToken, error) {
	var refreshToken domain.RefreshToken
	result := r.db.First(&refreshToken, "value = ", tokenValue)

	if result.Error != nil {
		return nil, result.Error
	}

	return &refreshToken, nil
}

func (r *postgresRefreshTokenRepo) GenerateTokenPair(userId uuid.UUID) (refreshToken string, accessToken string, err error) {
	unsignedRefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userId.String(),
	})

	unsignedAccessToken := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"userId":    userId.String(),
		"expiresAt": time.Now().Add(config.AccessTokenExpiresDuration),
	})

	refreshToken, err = unsignedRefreshToken.SignedString(os.Getenv("REFRESH_TOKEN_SECRET"))
	accessToken, err = unsignedAccessToken.SignedString(os.Getenv("ACCESS_TOKEN_SECRET"))

	if err != nil {
		return "", "", err
	}

	refreshTokenExpiresAt := time.Now().Add(config.RefreshTokenExpiresDuration)
	refreshTokenRecord, err := createRefreshTokenObject(refreshToken, refreshTokenExpiresAt)

	if err != nil {
		return "", "", err
	}

	if err := r.Create(*refreshTokenRecord); err != nil {
		return "", "", err
	}

	return
}

func createRefreshTokenObject(signedRefreshToken string, expiresAt time.Time) (*domain.RefreshToken, error) {
	uuid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	refreshToken := &domain.RefreshToken{
		ID:        uuid,
		Value:     signedRefreshToken,
		ExpiresAt: expiresAt,
	}

	return refreshToken, nil
}
