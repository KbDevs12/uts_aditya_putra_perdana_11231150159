package repository

import (
	"backend/internal/domain"

	"gorm.io/gorm"
)

type ProductRepo struct {
	db *gorm.DB
}

func NewProductRepo(db *gorm.DB) *ProductRepo {
	return &ProductRepo{db}
}

func (r *ProductRepo) FindAll() ([]domain.Product, error) {
	var products []domain.Product
	err := r.db.Order("id DESC").Find(&products).Error
	return products, err
}

func (r *ProductRepo) FindByID(id int64) (*domain.Product, error) {
	var product domain.Product
	err := r.db.First(&product, id).Error
	return &product, err
}
