package usecase

import (
	"backend/internal/domain"
	"backend/internal/repository"
)

type CartUsecase struct {
	cartRepo     *repository.CartRepo
	cartItemRepo *repository.CartItemRepo
}

func NewCartUsecase(cartRepo *repository.CartRepo, cartItemRepo *repository.CartItemRepo) *CartUsecase {
	return &CartUsecase{cartRepo, cartItemRepo}
}

func (u *CartUsecase) GetCart(userID int64) ([]domain.CartItem, error) {
	cart, _ := u.cartRepo.GetByUser(userID)
	if cart.ID == 0 {
		return []domain.CartItem{}, nil
	}
	return u.cartItemRepo.GetByCart(cart.ID)
}

func (u *CartUsecase) Add(userID, productID int64, qty int) error {
	cart, _ := u.cartRepo.GetByUser(userID)

	if cart.ID == 0 {
		cart = &domain.Cart{UserID: userID}
		u.cartRepo.Create(cart)
	}

	return u.cartItemRepo.Add(cart.ID, productID, qty)
}

func (u *CartUsecase) RemoveItem(cartItemID int64) error {
	return u.cartItemRepo.Remove(cartItemID)
}

func (u *CartUsecase) ClearCart(userID int64) error {
	cart, _ := u.cartRepo.GetByUser(userID)
	if cart.ID == 0 {
		return nil
	}
	return u.cartItemRepo.Clear(cart.ID)
}
