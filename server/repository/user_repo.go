package repository

import (
	"sen1or/lets-live/server/domain"

	"github.com/gofrs/uuid/v5"
	"gorm.io/gorm"
)

type postgresUserRepo struct {
	db gorm.DB
}

func NewUserRepository(conn gorm.DB) domain.UserRepository {
	return &postgresUserRepo{
		db: conn,
	}
}

func (r *postgresUserRepo) GetByID(userId uuid.UUID) (*domain.User, error) {
	var user domain.User
	result := r.db.First(&user, "id = ?", userId.String())
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (r *postgresUserRepo) GetByName(username string) (*domain.User, error) {
	var user domain.User
	result := r.db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (r *postgresUserRepo) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (r *postgresUserRepo) GetByAPIKey(apiKey uuid.UUID) (*domain.User, error) {
	var user domain.User
	result := r.db.Where("stream_api_key = ?", apiKey).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (r *postgresUserRepo) GetStreamingUsers() ([]domain.User, error) {
	var streamingUsers []domain.User
	result := r.db.Where("is_online = ?", true).Find(&streamingUsers)
	if result.Error != nil {
		return nil, result.Error
	}

	return streamingUsers, nil
}

func (r *postgresUserRepo) Create(newUser domain.User) error {
	result := r.db.Create(&newUser)
	return result.Error
}

// CAREFUL: gorm only update non-blank value (false, 0, nil...)
// if you want to update everything, use save
func (r *postgresUserRepo) Update(user domain.User) error {
	tx := r.db.Updates(&user)
	return tx.Error
}

func (r *postgresUserRepo) Save(user domain.User) error {
	tx := r.db.Save(&user)
	return tx.Error
}

func (r *postgresUserRepo) Delete(userId string) error {
	tx := r.db.Delete(&domain.User{}, userId)
	return tx.Error
}
