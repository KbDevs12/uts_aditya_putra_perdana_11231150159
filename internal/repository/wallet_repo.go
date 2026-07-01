package repository

import (
	"errors"

	"backend/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WalletRepo struct {
	db *gorm.DB
}

func NewWalletRepo(db *gorm.DB) *WalletRepo {
	return &WalletRepo{db: db}
}

func (r *WalletRepo) DB() *gorm.DB { return r.db }

func (r *WalletRepo) GetOrCreateWallet(userID int64) (*domain.WalletAccount, error) {
	var wallet domain.WalletAccount
	err := r.db.Where("user_id = ?", userID).First(&wallet).Error
	if err == nil {
		return &wallet, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	wallet = domain.WalletAccount{UserID: userID, Balance: 0}
	return &wallet, r.db.Create(&wallet).Error
}

func (r *WalletRepo) GetWalletForUpdate(tx *gorm.DB, userID int64) (*domain.WalletAccount, error) {
	var wallet domain.WalletAccount
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&wallet).Error
	return &wallet, err
}

func (r *WalletRepo) CreateTransaction(tx *gorm.DB, trx *domain.WalletTransaction) error {
	return tx.Create(trx).Error
}

func (r *WalletRepo) ListTransactions(userID int64, limit int) ([]domain.WalletTransaction, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var rows []domain.WalletTransaction
	err := r.db.Where("user_id = ?", userID).Order("id DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *WalletRepo) CreatePaymentIntent(intent *domain.PaymentIntent) error {
	return r.db.Create(intent).Error
}

func (r *WalletRepo) GetPaymentIntent(token string) (*domain.PaymentIntent, error) {
	var intent domain.PaymentIntent
	err := r.db.Where("token = ?", token).First(&intent).Error
	return &intent, err
}

func (r *WalletRepo) GetPaymentIntentForUpdate(tx *gorm.DB, token string) (*domain.PaymentIntent, error) {
	var intent domain.PaymentIntent
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("token = ?", token).First(&intent).Error
	return &intent, err
}

func (r *WalletRepo) SavePaymentIntent(intent *domain.PaymentIntent) error {
	return r.db.Save(intent).Error
}

func (r *WalletRepo) UpdatePINHash(userID int64, pinHash string) error {
	return r.db.Model(&domain.WalletAccount{}).Where("user_id = ?", userID).Update("pin_hash", pinHash).Error
}
