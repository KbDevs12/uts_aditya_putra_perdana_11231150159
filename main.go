package main

import (
	"context"
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
	redisClient := config.ConnectRedis()
	if redisClient != nil {
		defer redisClient.Close()
	}

	// Repositories
	userRepo := repository.NewUserRepo(db)
	productRepo := repository.NewProductRepo(db)
	cartRepo := repository.NewCartRepo(db)
	cartItemRepo := repository.NewCartItemRepo(db)
	orderRepo := repository.NewOrderRepo(db)
	orderItemRepo := repository.NewOrderItemRepo(db)
	walletRepo := repository.NewWalletRepo(db)

	// Usecases
	authUC := usecase.NewAuthUsecase(userRepo, redisClient)
	productUC := usecase.NewProductUsecase(productRepo)
	cartUC := usecase.NewCartUsecase(cartRepo, cartItemRepo)
	orderUC := usecase.NewOrderUsecase(orderRepo, orderItemRepo, cartRepo, cartItemRepo)
	walletUC := usecase.NewWalletUsecase(walletRepo, orderRepo, redisClient)

	// Handler
	h := deliveryHttp.NewHandler(authUC, productUC, cartUC, orderUC, walletUC)

	r := gin.Default()
	r.Use(corsMiddleware())

	r.GET("/health", func(c *gin.Context) {
		otpStorageStatus := "disabled"
		if redisClient != nil {
			if err := redisClient.Ping(context.Background()).Err(); err != nil {
				otpStorageStatus = "error"
			} else {
				otpStorageStatus = "ok"
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"otp_storage": otpStorageStatus,
		})
	})

	// Public routes
	r.POST("/auth/login", h.Login)
	r.POST("/auth/register", h.Register)
	r.POST("/auth/verify-email-otp", h.VerifyEmailOTP)
	r.POST("/otp/send-email", h.SendEmailOTP)
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
		api.POST("/wallet/transfer", h.TransferWallet)
		api.GET("/wallet/transactions", h.GetWalletTransactions)
		api.POST("/wallet/pin", h.SetWalletPIN)
		api.POST("/wallet/pin/verify", h.VerifyWalletPIN)
		api.POST("/payment-intents/:token/pay", h.PayPaymentIntent)

		// Account protection and notification endpoints
		api.POST("/auth/setup-2fa", h.SetupTwoFactor)
		api.POST("/auth/verify-2fa", h.VerifyTwoFactor)
		api.POST("/auth/notification-token", h.SaveNotificationToken)
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
