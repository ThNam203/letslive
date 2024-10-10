package repositories

import (
	"os"
	"sen1or/lets-live/api/config"
	"sen1or/lets-live/api/domains"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type postgresRefreshTokenRepo struct {
	db gorm.DB
}

func NewRefreshTokenRepository(conn gorm.DB) domains.RefreshTokenRepository {
	return &postgresRefreshTokenRepo{
		db: conn,
	}
}

func (r *postgresRefreshTokenRepo) RevokeByValue(tokenValue string) error {
	var refreshToken domains.RefreshToken
	result := r.db.First(&refreshToken, "value = ?", tokenValue)

	if result.Error != nil {
		return result.Error
	}

	result = r.db.Save(refreshToken)

	return result.Error
}

func (r *postgresRefreshTokenRepo) RevokeAll(userId uuid.UUID) error {
	var refreshToken domains.RefreshToken
	var timeNow = time.Now()
	result := r.db.Model(&refreshToken).Where("user_id = ?", userId).Updates(&domains.RefreshToken{RevokedAt: &timeNow})

	return result.Error
}
func (r *postgresRefreshTokenRepo) Create(tokenRecord domains.RefreshToken) error {
	result := r.db.Create(&tokenRecord)
	return result.Error
}

func (r *postgresRefreshTokenRepo) FindByValue(tokenValue string) (*domains.RefreshToken, error) {
	var refreshToken domains.RefreshToken
	result := r.db.First(&refreshToken, "value = ", tokenValue)

	if result.Error != nil {
		return nil, result.Error
	}

	return &refreshToken, nil
}

func (r *postgresRefreshTokenRepo) GenerateTokenPair(userId uuid.UUID) (refreshToken string, accessToken string, err error) {
	unsignedRefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":    userId.String(),
		"expiresAt": time.Now().Add(config.REFRESH_TOKEN_EXPIRES_DURATION),
	})

	unsignedAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":    userId.String(),
		"expiresAt": time.Now().Add(config.ACCESS_TOKEN_EXPIRES_DURATION),
	})

	refreshToken, err = unsignedRefreshToken.SignedString([]byte(os.Getenv("REFRESH_TOKEN_SECRET")))
	accessToken, err = unsignedAccessToken.SignedString([]byte(os.Getenv("ACCESS_TOKEN_SECRET")))

	if err != nil {
		return "", "", err
	}

	refreshTokenExpiresAt := time.Now().Add(config.REFRESH_TOKEN_EXPIRES_DURATION)
	refreshTokenRecord, err := createRefreshTokenObject(refreshToken, refreshTokenExpiresAt, userId)

	if err != nil {
		return "", "", err
	}

	if err := r.Create(*refreshTokenRecord); err != nil {
		return "", "", err
	}

	return
}

func createRefreshTokenObject(signedRefreshToken string, expiresAt time.Time, userId uuid.UUID) (*domains.RefreshToken, error) {
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
