package repository

import (
	"backend/internal/domain"
	"strings"
	"time"

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

func (r *UserRepo) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FindByID(id int64) (*domain.User, error) {
	var user domain.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) UpdateEmail(id int64, email string) error {
	return r.db.Model(&domain.User{}).Where("id = ?", id).Update("email", email).Error
}

func (r *UserRepo) Create(user *domain.User) error {
	if user != nil {
		user.TwoFactorMethod = NormalizeTwoFactorMethod(user.TwoFactorMethod)
	}
	return r.db.Create(user).Error
}

func (r *UserRepo) Save(user *domain.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepo) MarkEmailVerified(id int64) error {
	return r.db.Model(&domain.User{}).Where("id = ?", id).Updates(map[string]any{
		"email_verified": true,
	}).Error
}

func (r *UserRepo) UpdateTwoFactor(id int64, method string, enabled bool) error {
	return r.db.Model(&domain.User{}).Where("id = ?", id).Updates(map[string]any{
		"two_factor_method":  method,
		"two_factor_enabled": enabled,
	}).Error
}

func (r *UserRepo) SaveTOTPSecret(id int64, secret string) error {
	return r.db.Model(&domain.User{}).Where("id = ?", id).Updates(map[string]any{
		"totp_secret":        secret,
		"totp_enabled":       false,
		"totp_verified_at":   nil,
		"two_factor_method":  "totp",
		"two_factor_enabled": false,
	}).Error
}

func (r *UserRepo) MarkTOTPVerified(id int64, at time.Time) error {
	return r.db.Model(&domain.User{}).Where("id = ?", id).Updates(map[string]any{
		"two_factor_method":  "totp",
		"two_factor_enabled": true,
		"totp_enabled":       true,
		"totp_verified_at":   at,
	}).Error
}

func (r *UserRepo) UpdateNotificationToken(id int64, token string) error {
	return r.db.Model(&domain.User{}).Where("id = ?", id).Update("notification_token", token).Error
}

func (r *UserRepo) TouchLastLogin(id int64, at time.Time) error {
	return r.db.Model(&domain.User{}).Where("id = ?", id).Update("last_login_at", at).Error
}

func NormalizeTwoFactorMethod(method string) string {
	method = strings.TrimSpace(strings.ToLower(method))

	switch method {
	case "", "email", "email_otp", "smtp":
		return "smtp"
	case "authenticator", "google_authenticator", "totp":
		return "totp"
	case "notif", "notification", "push", "fcm":
		return "notif"
	default:
		return method
	}
}
