package http

import "backend/internal/usecase"

type Handler struct {
	authUC    *usecase.AuthUsecase
	productUC *usecase.ProductUsecase
	cartUC    *usecase.CartUsecase
	orderUC   *usecase.OrderUsecase
}
