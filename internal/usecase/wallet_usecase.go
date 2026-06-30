package usecase

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"backend/internal/domain"
	"backend/internal/repository"

	"gorm.io/gorm"
)

type WalletUsecase struct {
	walletRepo *repository.WalletRepo
	orderRepo  *repository.OrderRepo
}

func NewWalletUsecase(walletRepo *repository.WalletRepo, orderRepo *repository.OrderRepo) *WalletUsecase {
	return &WalletUsecase{walletRepo: walletRepo, orderRepo: orderRepo}
}

func (u *WalletUsecase) GetWallet(userID int64) (*domain.WalletAccount, error) {
	return u.walletRepo.GetOrCreateWallet(userID)
}

func (u *WalletUsecase) GetTransactions(userID int64, limit int) ([]domain.WalletTransaction, error) {
	return u.walletRepo.ListTransactions(userID, limit)
}

func (u *WalletUsecase) TopUp(userID int64, amount float64, description string) (*domain.WalletAccount, *domain.WalletTransaction, error) {
	if amount <= 0 {
		return nil, nil, errors.New("amount must be greater than 0")
	}
	if description == "" {
		description = "Top Up Saldo E-Wallet"
	}

	var wallet domain.WalletAccount
	var trx domain.WalletTransaction
	err := u.walletRepo.DB().Transaction(func(tx *gorm.DB) error {
		lockedWallet, err := u.walletRepo.GetWalletForUpdate(tx, userID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			lockedWallet = &domain.WalletAccount{UserID: userID, Balance: 0}
			if err := tx.Create(lockedWallet).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		before := lockedWallet.Balance
		lockedWallet.Balance += amount
		if err := tx.Save(lockedWallet).Error; err != nil {
			return err
		}

		trx = domain.WalletTransaction{
			WalletID:      lockedWallet.ID,
			UserID:        userID,
			ReferenceType: "topup",
			ReferenceID:   fmt.Sprintf("TOPUP-%d", time.Now().UnixNano()),
			Type:          domain.WalletTransactionCredit,
			Amount:        amount,
			Description:   description,
			BalanceBefore: before,
			BalanceAfter:  lockedWallet.Balance,
		}
		if err := u.walletRepo.CreateTransaction(tx, &trx); err != nil {
			return err
		}
		wallet = *lockedWallet
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return &wallet, &trx, nil
}

func (u *WalletUsecase) CreatePaymentIntent(orderID, userID int64, amount float64, merchantName string) (*domain.PaymentIntent, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}
	if merchantName == "" {
		merchantName = "E-Commerce Store"
	}

	token, err := generateToken(24)
	if err != nil {
		return nil, err
	}

	expires := time.Now().Add(30 * time.Minute)
	deepLink := buildWalletDeepLink(token, amount, merchantName)
	intent := &domain.PaymentIntent{
		Token:        token,
		OrderID:      orderID,
		UserID:       userID,
		MerchantName: merchantName,
		Amount:       amount,
		Status:       domain.PaymentIntentPending,
		DeepLink:     deepLink,
		ExpiresAt:    expires,
	}

	if err := u.walletRepo.CreatePaymentIntent(intent); err != nil {
		return nil, err
	}
	if err := u.orderRepo.AttachPaymentIntent(orderID, token); err != nil {
		return nil, err
	}
	return intent, nil
}

func (u *WalletUsecase) GetPaymentIntent(token string) (*domain.PaymentIntent, error) {
	intent, err := u.walletRepo.GetPaymentIntent(token)
	if err != nil {
		return nil, err
	}
	if intent.Status == domain.PaymentIntentPending && time.Now().After(intent.ExpiresAt) {
		intent.Status = domain.PaymentIntentExpired
	}
	return intent, nil
}

func (u *WalletUsecase) PayPaymentIntent(walletUserID int64, token string) (*domain.PaymentIntent, *domain.WalletTransaction, error) {
	var paidIntent domain.PaymentIntent
	var trx domain.WalletTransaction
	now := time.Now()

	err := u.walletRepo.DB().Transaction(func(tx *gorm.DB) error {
		intent, err := u.walletRepo.GetPaymentIntentForUpdate(tx, token)
		if err != nil {
			return errors.New("payment intent not found")
		}
		if intent.Status != domain.PaymentIntentPending {
			return errors.New("payment intent is not pending")
		}
		if time.Now().After(intent.ExpiresAt) {
			return errors.New("payment intent expired")
		}

		wallet, err := u.walletRepo.GetWalletForUpdate(tx, walletUserID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			wallet = &domain.WalletAccount{UserID: walletUserID, Balance: 0}
			if err := tx.Create(wallet).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		if wallet.Balance < intent.Amount {
			return errors.New("insufficient balance")
		}

		before := wallet.Balance
		wallet.Balance -= intent.Amount
		if err := tx.Save(wallet).Error; err != nil {
			return err
		}

		trx = domain.WalletTransaction{
			WalletID:      wallet.ID,
			UserID:        walletUserID,
			ReferenceType: "merchant_payment",
			ReferenceID:   intent.Token,
			Type:          domain.WalletTransactionDebit,
			Amount:        intent.Amount,
			Description:   fmt.Sprintf("Pembayaran %s - Order #%d", intent.MerchantName, intent.OrderID),
			BalanceBefore: before,
			BalanceAfter:  wallet.Balance,
		}
		if err := u.walletRepo.CreateTransaction(tx, &trx); err != nil {
			return err
		}

		intent.Status = domain.PaymentIntentPaid
		intent.PaidAt = &now
		if err := tx.Save(intent).Error; err != nil {
			return err
		}

		if err := u.orderRepo.MarkPaid(tx, intent.OrderID, "global_wallet", now); err != nil {
			return err
		}

		paidIntent = *intent
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return &paidIntent, &trx, nil
}

func generateToken(byteLen int) (string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func buildWalletDeepLink(token string, amount float64, merchantName string) string {
	scheme := os.Getenv("WALLET_DEEPLINK_SCHEME")
	if scheme == "" {
		scheme = "kantongin"
	}
	q := url.Values{}
	q.Set("token", token)
	q.Set("amount", fmt.Sprintf("%.0f", amount))
	q.Set("merchant", merchantName)
	return fmt.Sprintf("%s://pay?%s", scheme, q.Encode())
}
