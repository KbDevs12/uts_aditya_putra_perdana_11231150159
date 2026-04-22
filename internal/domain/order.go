package domain

type Order struct {
	ID         int64
	UserID     int64
	TotalPrice float64
	Status     string
}
