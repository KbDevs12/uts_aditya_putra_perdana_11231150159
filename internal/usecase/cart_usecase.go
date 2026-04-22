package usecase

import "backend/internal/domain"

func (u *CartUsecase) Add(userID, productID int64, qty int) error {
	cart, _ := u.cartRepo.GetByUser(userID)

	if cart.ID == 0 {
		cart = &domain.Cart{UserID: userID}
		u.cartRepo.Create(cart)
	}

	return u.cartItemRepo.Add(cart.ID, productID, qty)
}
