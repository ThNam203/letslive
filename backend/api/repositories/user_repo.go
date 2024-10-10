package repositories

import (
	"sen1or/lets-live/api/domains"

	"github.com/gofrs/uuid/v5"
	"gorm.io/gorm"
)

type postgresUserRepo struct {
	db gorm.DB
}

func NewUserRepository(conn gorm.DB) domains.UserRepository {
	return &postgresUserRepo{
		db: conn,
	}
}

func (r *postgresUserRepo) GetByID(userId uuid.UUID) (*domains.User, error) {
	var user domains.User
	result := r.db.First(&user, "id = ?", userId.String())
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (r *postgresUserRepo) GetByName(username string) (*domains.User, error) {
	var user domains.User
	result := r.db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (r *postgresUserRepo) GetByEmail(email string) (*domains.User, error) {
	var user domains.User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (r *postgresUserRepo) GetByAPIKey(apiKey uuid.UUID) (*domains.User, error) {
	var user domains.User
	result := r.db.Where("stream_api_key = ?", apiKey).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (r *postgresUserRepo) GetStreamingUsers() ([]domains.User, error) {
	var streamingUsers []domains.User
	result := r.db.Where("is_online = ?", true).Find(&streamingUsers)
	if result.Error != nil {
		return nil, result.Error
	}

	return streamingUsers, nil
}

func (r *postgresUserRepo) Create(newUser domains.User) error {
	result := r.db.Create(&newUser)
	return result.Error
}

// CAREFUL: gorm only update non-blank value (false, 0, nil...)
// if you want to update everything, use save
func (r *postgresUserRepo) Update(user domains.User) error {
	tx := r.db.Updates(&user)
	return tx.Error
}

func (r *postgresUserRepo) Save(user domains.User) error {
	tx := r.db.Save(&user)
	return tx.Error
}

func (r *postgresUserRepo) Delete(userId string) error {
	tx := r.db.Delete(&domains.User{}, userId)
	return tx.Error
}
