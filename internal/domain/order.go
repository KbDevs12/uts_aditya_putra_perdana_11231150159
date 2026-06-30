package domain

import "time"

type Order struct {
	ID                 int64      `gorm:"primaryKey" json:"id"`
	UserID             int64      `json:"user_id"`
	TotalPrice         float64    `json:"total_price"`
	Status             string     `json:"status"`
	PaymentStatus      string     `gorm:"column:payment_status;default:unpaid" json:"payment_status"`
	PaymentMethod      string     `gorm:"column:payment_method" json:"payment_method"`
	PaymentIntentToken string     `gorm:"column:payment_intent_token;index" json:"payment_intent_token,omitempty"`
	PaidAt             *time.Time `gorm:"column:paid_at" json:"paid_at,omitempty"`
	CreatedAt          time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (Order) TableName() string { return "orders" }

type OrderItem struct {
	ID        int64   `gorm:"primaryKey" json:"id"`
	OrderID   int64   `json:"order_id"`
	ProductID int64   `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

func (OrderItem) TableName() string { return "order_items" }
