package config

import (
	"log"
	"os"

	"backend/internal/domain"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() *gorm.DB {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN is empty. Set DB_DSN in .env")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(
		&domain.User{},
		&domain.Product{},
		&domain.Cart{},
		&domain.CartItem{},
		&domain.Order{},
		&domain.OrderItem{},
		&domain.WalletAccount{},
		&domain.WalletTransaction{},
		&domain.PaymentIntent{},
	); err != nil {
		log.Fatal("failed to migrate database: ", err)
	}

	return db
}
