package domain

type CartItem struct {
	ID        int64   `gorm:"primaryKey" json:"id"`
	CartID    int64   `json:"cart_id"`
	ProductID int64   `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

func (CartItem) TableName() string { return "cart_items" }
