package usecase

import (
	"backend/internal/domain"
	"backend/internal/repository"
)

type ProductUsecase struct {
	repo *repository.ProductRepo
}

func NewProductUsecase(repo *repository.ProductRepo) *ProductUsecase {
	return &ProductUsecase{repo}
}

func (u *ProductUsecase) GetAll() ([]domain.Product, error) {
	return u.repo.FindAll()
}

func (u *ProductUsecase) GetByID(id int64) (*domain.Product, error) {
	return u.repo.FindByID(id)
}
