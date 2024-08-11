package repository

import (
	"errors"
	"sen1or/lets-live/server/domain"

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

func (r *postgresUserRepo) GetByID(userId string) (*domain.User, error) {
	var user domain.User
	result := r.db.First(&user, "id = ?", userId)
	if result.Error != nil {
		if errors.Is(gorm.ErrRecordNotFound, result.Error) {
			return nil, gorm.ErrRecordNotFound
		} else {
			return nil, errors.New("internal server error")
		}
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

func (r *postgresUserRepo) Create(newUser domain.User) error {
	result := r.db.Create(newUser)
	return result.Error
}

func (r *postgresUserRepo) Update(user domain.User) error {
	tx := r.db.Model(&user).Where("id = ", user.ID).Updates(&user)
	return tx.Error

}

func (r *postgresUserRepo) Delete(userId string) error {
	tx := r.db.Delete(&domain.User{}, userId)
	return tx.Error
}
