package domain

type Cart struct {
	ID     int64 `gorm:"primaryKey" json:"id"`
	UserID int64 `json:"user_id"`
}

func (Cart) TableName() string { return "carts" }
