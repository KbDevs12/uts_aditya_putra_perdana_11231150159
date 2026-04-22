package usecase

import (
	"errors"

	"backend/internal/domain"
	"backend/internal/repository"
)

type OrderUsecase struct {
	orderRepo     *repository.OrderRepo
	orderItemRepo *repository.OrderItemRepo
	cartRepo      *repository.CartRepo
	cartItemRepo  *repository.CartItemRepo
}

func NewOrderUsecase(
	orderRepo *repository.OrderRepo,
	orderItemRepo *repository.OrderItemRepo,
	cartRepo *repository.CartRepo,
	cartItemRepo *repository.CartItemRepo,
) *OrderUsecase {
	return &OrderUsecase{orderRepo, orderItemRepo, cartRepo, cartItemRepo}
}

func (u *OrderUsecase) Checkout(userID int64) (*domain.Order, error) {
	cart, _ := u.cartRepo.GetByUser(userID)
	if cart.ID == 0 {
		return nil, errors.New("cart is empty")
	}

	items, err := u.cartItemRepo.GetByCart(cart.ID)
	if err != nil || len(items) == 0 {
		return nil, errors.New("cart is empty")
	}

	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}

	order := &domain.Order{
		UserID:     userID,
		TotalPrice: total,
		Status:     "pending",
	}
	if err := u.orderRepo.Create(order); err != nil {
		return nil, err
	}

	var orderItems []domain.OrderItem
	for _, item := range items {
		orderItems = append(orderItems, domain.OrderItem{
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}
	if err := u.orderItemRepo.CreateBulk(orderItems); err != nil {
		return nil, err
	}

	u.cartItemRepo.Clear(cart.ID)

	return order, nil
}

func (u *OrderUsecase) GetMyOrders(userID int64) ([]domain.Order, error) {
	return u.orderRepo.GetByUser(userID)
}

func (u *OrderUsecase) GetOrderDetail(orderID, userID int64) (*domain.Order, []domain.OrderItem, error) {
	order, err := u.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, nil, errors.New("order not found")
	}
	if order.UserID != userID {
		return nil, nil, errors.New("forbidden")
	}
	items, err := u.orderItemRepo.GetByOrder(orderID)
	return order, items, err
}
