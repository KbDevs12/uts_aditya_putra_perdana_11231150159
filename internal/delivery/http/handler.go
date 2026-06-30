package http

import (
	"backend/internal/usecase"
)

type Handler struct {
	authUC    *usecase.AuthUsecase
	productUC *usecase.ProductUsecase
	cartUC    *usecase.CartUsecase
	orderUC   *usecase.OrderUsecase
	walletUC  *usecase.WalletUsecase
}

func NewHandler(
	authUC *usecase.AuthUsecase,
	productUC *usecase.ProductUsecase,
	cartUC *usecase.CartUsecase,
	orderUC *usecase.OrderUsecase,
	walletUC *usecase.WalletUsecase,
) *Handler {
	return &Handler{authUC, productUC, cartUC, orderUC, walletUC}
}
