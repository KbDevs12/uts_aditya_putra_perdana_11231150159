package repository

import (
	"backend/internal/domain"

	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db}
}

func (r *UserRepo) FindByUID(uid string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("firebase_uid = ?", uid).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) UpdateEmail(id int64, email string) error {
	return r.db.Model(&domain.User{}).Where("id = ?", id).Update("email", email).Error
}

func (r *UserRepo) Create(user *domain.User) error {
	return r.db.Create(user).Error
}
