package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"backend/config"
	deliveryHttp "backend/internal/delivery/http"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/usecase"

	"golang.ngrok.com/ngrok/v2"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config.InitFirebase()
	db := config.ConnectDB()

	// Repositories
	userRepo := repository.NewUserRepo(db)
	productRepo := repository.NewProductRepo(db)
	cartRepo := repository.NewCartRepo(db)
	cartItemRepo := repository.NewCartItemRepo(db)
	orderRepo := repository.NewOrderRepo(db)
	orderItemRepo := repository.NewOrderItemRepo(db)

	authUC := usecase.NewAuthUsecase(userRepo)
	productUC := usecase.NewProductUsecase(productRepo)
	cartUC := usecase.NewCartUsecase(cartRepo, cartItemRepo)
	orderUC := usecase.NewOrderUsecase(orderRepo, orderItemRepo, cartRepo, cartItemRepo)

	// Handler
	h := deliveryHttp.NewHandler(authUC, productUC, cartUC, orderUC)

	r := gin.Default()

	// Public routes
	r.POST("/auth/login", h.Login)
	r.POST("/auth/register", h.Register)

	// Protected routes
	api := r.Group("/api", middleware.JWTAuth())
	{
		// Products
		api.GET("/products", h.GetProducts)
		api.GET("/products/:id", h.GetDetail)

		// Cart
		api.GET("/cart", h.GetCart)
		api.POST("/cart", h.AddToCart)
		api.DELETE("/cart/:id", h.RemoveFromCart)
		api.DELETE("/cart", h.ClearCart)

		// Orders
		api.POST("/orders/checkout", h.Checkout)
		api.GET("/orders", h.GetMyOrders)
		api.GET("/orders/:id", h.GetOrderDetail)
	}

	ctx := context.Background()
	tunnel, err := ngrok.Listen(ctx)
	if err != nil {
		log.Fatal("failed to start ngrok", err)
	}

	log.Println("public url: ", tunnel.URL())

	if err := http.Serve(tunnel, r); err != nil {
		log.Fatal("Server error:", err)
	}
}
