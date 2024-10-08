package repository

import (
	"log"
	"sen1or/lets-live/server/domain"
	"time"

	"github.com/gofrs/uuid/v5"
	"gorm.io/gorm"
)

type postgresVerifyTokenRepo struct {
	db gorm.DB
}

func NewVerifyTokenRepo(conn gorm.DB) domain.VerifyTokenRepository {
	return &postgresVerifyTokenRepo{
		db: conn,
	}
}

func (r *postgresVerifyTokenRepo) CreateToken(userId uuid.UUID) (*domain.VerifyToken, error) {
	token, _ := uuid.NewGen().NewV4()
	newToken := &domain.VerifyToken{
		Token:     token.String(),
		ExpiresAt: time.Now().Add(1 * time.Hour), // Token validity
		UserID:    userId,
	}

	result := r.db.Create(newToken)
	if result.Error != nil {
		log.Printf("error while creating new verify token: %s", result.Error.Error())
	}

	return newToken, nil
}

func (r *postgresVerifyTokenRepo) GetByToken(token string) (*domain.VerifyToken, error) {
	var record domain.VerifyToken
	result := r.db.First(&record, "token = ?", token)

	if result.Error != nil {
		return nil, result.Error
	}

	return &record, nil
}

func (r *postgresVerifyTokenRepo) UpdateToken(verifyToken domain.VerifyToken) error {
	result := r.db.Model(&verifyToken).Updates(&verifyToken)
	return result.Error
}

func (r *postgresVerifyTokenRepo) DeleteToken(tokenId uuid.UUID) error {
	result := r.db.Where("id = ?", tokenId).Delete(&domain.VerifyToken{})
	return result.Error
}
