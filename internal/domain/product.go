package domain

type Product struct {
	ID             int64   `gorm:"primaryKey" json:"id"`
	Name           string  `json:"name"`
	Brand          string  `json:"brand"`
	Description    string  `json:"description"`
	Price          float64 `json:"price"`
	Stock          int     `json:"stock"`
	ImageURL       string  `gorm:"column:image_url" json:"image_url"`
	Type           string  `json:"type"`
	LongevityHours int     `gorm:"column:longevity_hours" json:"longevity_hours"`
}

func (Product) TableName() string { return "products" }
