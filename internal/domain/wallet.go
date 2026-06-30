package domain

import "time"

const (
	WalletTransactionCredit = "credit"
	WalletTransactionDebit  = "debit"

	PaymentIntentPending = "pending"
	PaymentIntentPaid    = "paid"
	PaymentIntentExpired = "expired"
	PaymentIntentFailed  = "failed"
)

type WalletAccount struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	UserID    int64     `gorm:"uniqueIndex;not null" json:"user_id"`
	Balance   float64   `gorm:"type:numeric(15,2);default:0" json:"balance"`
	PINHash   string    `gorm:"column:pin_hash" json:"-"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (WalletAccount) TableName() string { return "wallet_accounts" }

type WalletTransaction struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	WalletID      int64     `gorm:"column:wallet_id;index" json:"wallet_id"`
	UserID        int64     `gorm:"column:user_id;index" json:"user_id"`
	ReferenceType string    `gorm:"column:reference_type" json:"reference_type"`
	ReferenceID   string    `gorm:"column:reference_id;index" json:"reference_id"`
	Type          string    `gorm:"type:varchar(20)" json:"type"`
	Amount        float64   `gorm:"type:numeric(15,2)" json:"amount"`
	Description   string    `gorm:"type:varchar(255)" json:"description"`
	BalanceBefore float64   `gorm:"type:numeric(15,2)" json:"balance_before"`
	BalanceAfter  float64   `gorm:"type:numeric(15,2)" json:"balance_after"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
}

func (WalletTransaction) TableName() string { return "wallet_transactions" }

type PaymentIntent struct {
	ID           int64      `gorm:"primaryKey" json:"id"`
	Token        string     `gorm:"uniqueIndex;not null" json:"token"`
	OrderID      int64      `gorm:"column:order_id;index" json:"order_id"`
	UserID       int64      `gorm:"column:user_id;index" json:"user_id"`
	MerchantName string     `gorm:"column:merchant_name" json:"merchant_name"`
	Amount       float64    `gorm:"type:numeric(15,2)" json:"amount"`
	Status       string     `gorm:"type:varchar(30);default:pending" json:"status"`
	DeepLink     string     `gorm:"column:deep_link;type:text" json:"deep_link"`
	ExpiresAt    time.Time  `gorm:"column:expires_at" json:"expires_at"`
	PaidAt       *time.Time `gorm:"column:paid_at" json:"paid_at,omitempty"`
	CreatedAt    time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (PaymentIntent) TableName() string { return "payment_intents" }
