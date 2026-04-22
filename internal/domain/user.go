package domain

type User struct {
	ID          int64  `gorm:"primaryKey" json:"id"`
	FirebaseUID string `gorm:"column:firebase_uid;uniqueIndex" json:"-"`
	Name        string `json:"name"`
	Email       string `json:"email"`
}

func (User) TableName() string { return "users" }
