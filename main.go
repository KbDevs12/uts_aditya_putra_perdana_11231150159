package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

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

	token := os.Getenv("NGROK_AUTHTOKEN")
	if token == "" {
		log.Fatal("❌ NGROK_AUTHTOKEN tidak ditemukan di .env!")
	}
	log.Println("✅ NGROK_AUTHTOKEN ditemukan:", token[:8]+"...")

	l, err := ngrok.Listen(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("🌐 Public URL:", l.URL())

	if err := http.Serve(l, r); err != nil {
		log.Fatal(err)
	}
}
