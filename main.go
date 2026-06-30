package main

import (
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
	walletRepo := repository.NewWalletRepo(db)

	// Usecases
	authUC := usecase.NewAuthUsecase(userRepo)
	productUC := usecase.NewProductUsecase(productRepo)
	cartUC := usecase.NewCartUsecase(cartRepo, cartItemRepo)
	orderUC := usecase.NewOrderUsecase(orderRepo, orderItemRepo, cartRepo, cartItemRepo)
	walletUC := usecase.NewWalletUsecase(walletRepo, orderRepo)

	// Handler
	h := deliveryHttp.NewHandler(authUC, productUC, cartUC, orderUC, walletUC)

	r := gin.Default()
	r.Use(corsMiddleware())

	// Public routes
	r.POST("/auth/login", h.Login)
	r.POST("/auth/register", h.Register)
	r.GET("/api/payment-intents/:token", h.GetPaymentIntent)

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
		api.POST("/orders/:id/payment-intent", h.CreateOrderPaymentIntent)

		// E-wallet
		api.GET("/wallet", h.GetWallet)
		api.POST("/wallet/topup", h.TopUpWallet)
		api.GET("/wallet/transactions", h.GetWalletTransactions)
		api.POST("/payment-intents/:token/pay", h.PayPaymentIntent)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Backend running on http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
