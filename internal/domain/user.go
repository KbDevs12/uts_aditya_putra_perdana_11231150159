package domain

import "time"

type User struct {
	ID               int64      `gorm:"primaryKey" json:"id"`
	FirebaseUID      string     `gorm:"column:firebase_uid;uniqueIndex" json:"-"`
	Name             string     `json:"name"`
	Email            string     `gorm:"uniqueIndex" json:"email"`
	EmailVerified    bool       `gorm:"column:email_verified;default:false" json:"email_verified"`
	TwoFactorMethod  string     `gorm:"column:two_factor_method;default:smtp" json:"two_factor_method"`
	TwoFactorEnabled bool       `gorm:"column:two_factor_enabled;default:false" json:"two_factor_enabled"`
	TOTPSecret       string     `gorm:"column:totp_secret;type:text" json:"-"`
	TOTPEnabled      bool       `gorm:"column:totp_enabled;default:false" json:"totp_enabled"`
	TOTPVerifiedAt   *time.Time `gorm:"column:totp_verified_at" json:"totp_verified_at,omitempty"`
	FCMToken         string     `gorm:"column:fcm_token;type:text" json:"-"`
	CreatedAt        time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at" json:"updated_at"`
	LastLoginAt      *time.Time `gorm:"column:last_login_at" json:"last_login_at,omitempty"`
}

func (User) TableName() string { return "users" }
