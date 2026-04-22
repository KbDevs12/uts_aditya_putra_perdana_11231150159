package repository

import (
	"backend/internal/domain"

	"gorm.io/gorm"
)

type CartRepo struct {
	db *gorm.DB
}

func NewCartRepo(db *gorm.DB) *CartRepo {
	return &CartRepo{db}
}

func (r *CartRepo) GetByUser(userID int64) (*domain.Cart, error) {
	var cart domain.Cart
	err := r.db.Where("user_id = ?", userID).First(&cart).Error
	if err != nil {
		return &domain.Cart{}, nil
	}
	return &cart, nil
}

func (r *CartRepo) Create(cart *domain.Cart) error {
	return r.db.Create(cart).Error
}

type CartItemRepo struct {
	db *gorm.DB
}

func NewCartItemRepo(db *gorm.DB) *CartItemRepo {
	return &CartItemRepo{db}
}

func (r *CartItemRepo) GetByCart(cartID int64) ([]domain.CartItem, error) {
	var items []domain.CartItem
	err := r.db.Where("cart_id = ?", cartID).Find(&items).Error
	return items, err
}

func (r *CartItemRepo) Add(cartID, productID int64, qty int) error {
	// Get product price
	var product struct{ Price float64 }
	if err := r.db.Table("products").Select("price").Where("id = ?", productID).Scan(&product).Error; err != nil {
		return err
	}

	// Upsert: if item already exists, increase qty
	var existing domain.CartItem
	err := r.db.Where("cart_id = ? AND product_id = ?", cartID, productID).First(&existing).Error
	if err == nil {
		return r.db.Model(&existing).Update("quantity", existing.Quantity+qty).Error
	}

	item := domain.CartItem{
		CartID:    cartID,
		ProductID: productID,
		Quantity:  qty,
		Price:     product.Price,
	}
	return r.db.Create(&item).Error
}

func (r *CartItemRepo) Remove(cartItemID int64) error {
	return r.db.Delete(&domain.CartItem{}, cartItemID).Error
}

func (r *CartItemRepo) Clear(cartID int64) error {
	return r.db.Where("cart_id = ?", cartID).Delete(&domain.CartItem{}).Error
}
