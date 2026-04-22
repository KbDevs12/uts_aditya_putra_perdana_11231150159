package domain

type Product struct {
	ID             int64
	Name           string
	Brand          string
	Description    string
	Price          float64
	Stock          int
	ImageURL       string
	Type           string
	LongevityHours int
}
