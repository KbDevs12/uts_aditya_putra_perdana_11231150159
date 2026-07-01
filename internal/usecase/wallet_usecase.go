package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"backend/internal/domain"
	"backend/internal/repository"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type WalletUsecase struct {
	walletRepo  *repository.WalletRepo
	orderRepo   *repository.OrderRepo
	redisClient *redis.Client
}

const paymentIntentCachePrefix = "wallet:payment_intent:"

func NewWalletUsecase(walletRepo *repository.WalletRepo, orderRepo *repository.OrderRepo, redisClient *redis.Client) *WalletUsecase {
	return &WalletUsecase{walletRepo: walletRepo, orderRepo: orderRepo, redisClient: redisClient}
}

func (u *WalletUsecase) GetWallet(userID int64) (*domain.WalletAccount, error) {
	return u.walletRepo.GetOrCreateWallet(userID)
}

func (u *WalletUsecase) GetTransactions(userID int64, limit int) ([]domain.WalletTransaction, error) {
	return u.walletRepo.ListTransactions(userID, limit)
}

func (u *WalletUsecase) SetPIN(userID int64, pin string) error {
	if len(pin) != 6 {
		return errors.New("pin must be exactly 6 digits")
	}
	for _, ch := range pin {
		if ch < '0' || ch > '9' {
			return errors.New("pin must contain digits only")
		}
	}
	if _, err := u.walletRepo.GetOrCreateWallet(userID); err != nil {
		return err
	}
	pinHash := hashPIN(pin)
	return u.walletRepo.UpdatePINHash(userID, pinHash)
}

func (u *WalletUsecase) VerifyPIN(userID int64, pin string) error {
	wallet, err := u.walletRepo.GetOrCreateWallet(userID)
	if err != nil {
		return err
	}
	if wallet.PINHash == "" {
		return errors.New("pin setup required")
	}
	if strings.TrimSpace(pin) == "" {
		return errors.New("pin is required")
	}
	if wallet.PINHash != hashPIN(pin) {
		return errors.New("invalid pin")
	}
	return nil
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
	u.cachePaymentIntent(context.Background(), intent)
	return intent, nil
}

func (u *WalletUsecase) GetPaymentIntent(token string) (*domain.PaymentIntent, error) {
	intent, err := u.walletRepo.GetPaymentIntent(token)
	if err != nil {
		return nil, err
	}
	if intent.Status == domain.PaymentIntentPending && time.Now().After(intent.ExpiresAt) {
		intent.Status = domain.PaymentIntentExpired
		_ = u.walletRepo.SavePaymentIntent(intent)
		u.deletePaymentIntentCache(context.Background(), token)
	}
	return intent, nil
}

func (u *WalletUsecase) PayPaymentIntent(walletUserID int64, token string, pin string) (*domain.PaymentIntent, *domain.WalletTransaction, error) {
	var paidIntent domain.PaymentIntent
	var trx domain.WalletTransaction
	now := time.Now()

	err := u.walletRepo.DB().Transaction(func(tx *gorm.DB) error {
		intent, err := u.walletRepo.GetPaymentIntentForUpdate(tx, token)
		if err != nil {
			return errors.New("payment intent not found")
		}
		if intent.UserID != walletUserID {
			return errors.New("payment intent does not belong to this account")
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
		if wallet.PINHash == "" {
			return errors.New("pin setup required")
		}
		if strings.TrimSpace(pin) == "" {
			return errors.New("pin is required")
		}
		if wallet.PINHash != hashPIN(pin) {
			return errors.New("invalid pin")
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

		if err := u.orderRepo.MarkPaid(tx, intent.OrderID, "kantongin", now); err != nil {
			return err
		}

		paidIntent = *intent
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	u.deletePaymentIntentCache(context.Background(), token)
	return &paidIntent, &trx, nil
}

func (u *WalletUsecase) Transfer(senderID int64, receiverID int64, receiverLabel string, amount float64, pin string) (*domain.WalletTransaction, *domain.WalletTransaction, error) {
	if senderID == receiverID {
		return nil, nil, errors.New("receiver must be different from sender")
	}
	if amount <= 0 {
		return nil, nil, errors.New("amount must be greater than 0")
	}
	if strings.TrimSpace(pin) == "" {
		return nil, nil, errors.New("pin is required")
	}

	var debitTrx domain.WalletTransaction
	var creditTrx domain.WalletTransaction
	referenceID := fmt.Sprintf("TRF-%d", time.Now().UnixNano())

	err := u.walletRepo.DB().Transaction(func(tx *gorm.DB) error {
		senderWallet, err := u.walletRepo.GetWalletForUpdate(tx, senderID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("wallet not found")
		}
		if err != nil {
			return err
		}
		if senderWallet.PINHash == "" {
			return errors.New("pin setup required")
		}
		if senderWallet.PINHash != hashPIN(pin) {
			return errors.New("invalid pin")
		}
		if senderWallet.Balance < amount {
			return errors.New("insufficient balance")
		}

		receiverWallet, err := u.walletRepo.GetWalletForUpdate(tx, receiverID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			receiverWallet = &domain.WalletAccount{UserID: receiverID, Balance: 0}
			if err := tx.Create(receiverWallet).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		senderBefore := senderWallet.Balance
		receiverBefore := receiverWallet.Balance
		senderWallet.Balance -= amount
		receiverWallet.Balance += amount

		if err := tx.Save(senderWallet).Error; err != nil {
			return err
		}
		if err := tx.Save(receiverWallet).Error; err != nil {
			return err
		}

		debitTrx = domain.WalletTransaction{
			WalletID:      senderWallet.ID,
			UserID:        senderID,
			ReferenceType: "transfer_out",
			ReferenceID:   referenceID,
			Type:          domain.WalletTransactionDebit,
			Amount:        amount,
			Description:   fmt.Sprintf("Transfer ke %s", receiverLabel),
			BalanceBefore: senderBefore,
			BalanceAfter:  senderWallet.Balance,
		}
		if err := u.walletRepo.CreateTransaction(tx, &debitTrx); err != nil {
			return err
		}

		creditTrx = domain.WalletTransaction{
			WalletID:      receiverWallet.ID,
			UserID:        receiverID,
			ReferenceType: "transfer_in",
			ReferenceID:   referenceID,
			Type:          domain.WalletTransactionCredit,
			Amount:        amount,
			Description:   "Terima transfer Kantongin",
			BalanceBefore: receiverBefore,
			BalanceAfter:  receiverWallet.Balance,
		}
		return u.walletRepo.CreateTransaction(tx, &creditTrx)
	})
	if err != nil {
		return nil, nil, err
	}
	return &debitTrx, &creditTrx, nil
}

func hashPIN(pin string) string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "kantongin-secret"
	}
	sum := sha256.Sum256([]byte(secret + ":" + pin))
	return hex.EncodeToString(sum[:])
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

func paymentIntentCacheKey(token string) string {
	return paymentIntentCachePrefix + token
}

func (u *WalletUsecase) cachePaymentIntent(ctx context.Context, intent *domain.PaymentIntent) {
	if u.redisClient == nil || intent == nil || intent.Token == "" {
		return
	}

	ttl := time.Until(intent.ExpiresAt)
	if ttl <= 0 {
		return
	}

	payload, err := json.Marshal(intent)
	if err != nil {
		return
	}

	_ = u.redisClient.Set(ctx, paymentIntentCacheKey(intent.Token), payload, ttl).Err()
}

func (u *WalletUsecase) deletePaymentIntentCache(ctx context.Context, token string) {
	if u.redisClient == nil || token == "" {
		return
	}
	_ = u.redisClient.Del(ctx, paymentIntentCacheKey(token)).Err()
}
