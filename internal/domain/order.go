package domain

type Order struct {
	ID         int64   `gorm:"primaryKey" json:"id"`
	UserID     int64   `json:"user_id"`
	TotalPrice float64 `json:"total_price"`
	Status     string  `json:"status"`
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
