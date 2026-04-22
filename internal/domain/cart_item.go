package domain

type CartItem struct {
	ID        int64
	CartID    int64
	ProductID int64
	Quantity  int
	Price     float64
}
