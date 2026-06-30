package repository

import (
	"time"

	"backend/internal/domain"

	"gorm.io/gorm"
)

type OrderRepo struct {
	db *gorm.DB
}

func NewOrderRepo(db *gorm.DB) *OrderRepo {
	return &OrderRepo{db}
}

func (r *OrderRepo) Create(order *domain.Order) error {
	return r.db.Create(order).Error
}

func (r *OrderRepo) GetByUser(userID int64) ([]domain.Order, error) {
	var orders []domain.Order
	err := r.db.Where("user_id = ?", userID).Order("id DESC").Find(&orders).Error
	return orders, err
}

func (r *OrderRepo) GetByID(id int64) (*domain.Order, error) {
	var order domain.Order
	err := r.db.First(&order, id).Error
	return &order, err
}

func (r *OrderRepo) AttachPaymentIntent(orderID int64, token string) error {
	return r.db.Model(&domain.Order{}).Where("id = ?", orderID).Updates(map[string]interface{}{
		"payment_intent_token": token,
		"payment_status":       "unpaid",
		"payment_method":       "global_wallet",
	}).Error
}

func (r *OrderRepo) MarkPaid(tx *gorm.DB, orderID int64, paymentMethod string, paidAt time.Time) error {
	return tx.Model(&domain.Order{}).Where("id = ?", orderID).Updates(map[string]interface{}{
		"status":         "paid",
		"payment_status": "paid",
		"payment_method": paymentMethod,
		"paid_at":        paidAt,
	}).Error
}

type OrderItemRepo struct {
	db *gorm.DB
}

func NewOrderItemRepo(db *gorm.DB) *OrderItemRepo {
	return &OrderItemRepo{db}
}

func (r *OrderItemRepo) CreateBulk(items []domain.OrderItem) error {
	return r.db.Create(&items).Error
}

func (r *OrderItemRepo) GetByOrder(orderID int64) ([]domain.OrderItem, error) {
	var items []domain.OrderItem
	err := r.db.Where("order_id = ?", orderID).Find(&items).Error
	return items, err
}
