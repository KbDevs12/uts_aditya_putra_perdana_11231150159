package http

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type topUpWalletRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
}

func (h *Handler) GetWallet(c *gin.Context) {
	userID := c.GetInt64("user_id")
	email := c.GetString("email")
	wallet, err := h.walletUC.GetWallet(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ownerName := "Pengguna Kantongin"
	if email != "" && strings.Contains(email, "@") {
		ownerName = strings.Title(strings.Split(email, "@")[0])
	}
	c.JSON(http.StatusOK, gin.H{
		"id":                    wallet.ID,
		"user_id":               wallet.UserID,
		"balance":               wallet.Balance,
		"created_at":            wallet.CreatedAt,
		"updated_at":            wallet.UpdatedAt,
		"verified":              true,
		"owner_name":            ownerName,
		"email":                 email,
		"active_checkout_count": 0,
		"saved_payment_count":   1,
		"pin_enabled":           wallet.PINHash != "",
	})
}

func (h *Handler) TopUpWallet(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req topUpWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount is required and must be greater than 0"})
		return
	}
	if req.Description == "" {
		req.Description = "Top up saldo Kantongin"
	}
	wallet, trx, err := h.walletUC.TopUp(userID, req.Amount, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"wallet": wallet, "transaction": trx})
}

func (h *Handler) GetWalletTransactions(c *gin.Context) {
	userID := c.GetInt64("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	rows, err := h.walletUC.GetTransactions(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (h *Handler) GetPaymentIntent(c *gin.Context) {
	token := c.Param("token")
	intent, err := h.walletUC.GetPaymentIntent(token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment intent not found"})
		return
	}
	c.JSON(http.StatusOK, intent)
}

func (h *Handler) SetWalletPIN(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req struct {
		PIN string `json:"pin" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pin is required"})
		return
	}
	if err := h.walletUC.SetPIN(userID, req.PIN); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "pin saved"})
}

func (h *Handler) VerifyWalletPIN(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req struct {
		PIN string `json:"pin" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pin is required"})
		return
	}
	if err := h.walletUC.VerifyPIN(userID, req.PIN); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "pin valid"})
}

func (h *Handler) PayPaymentIntent(c *gin.Context) {
	userID := c.GetInt64("user_id")
	token := c.Param("token")
	var req struct {
		PIN string `json:"pin"`
	}
	_ = c.ShouldBindJSON(&req)
	intent, trx, err := h.walletUC.PayPaymentIntent(userID, token, req.PIN)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "insufficient balance" {
			status = http.StatusPaymentRequired
		} else if err.Error() == "invalid pin" || err.Error() == "pin setup required" || err.Error() == "pin is required" {
			status = http.StatusUnauthorized
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":        "payment successful",
		"payment_intent": intent,
		"transaction":    trx,
	})
}

func (h *Handler) TransferWallet(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req struct {
		ReceiverEmail string  `json:"receiver_email" binding:"required"`
		Amount        float64 `json:"amount" binding:"required,gt=0"`
		PIN           string  `json:"pin" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "receiver email, amount, and pin are required"})
		return
	}
	receiver, err := h.authUC.FindUserByEmail(req.ReceiverEmail)
	if err != nil || receiver == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "receiver account not found"})
		return
	}
	label := receiver.Email
	if receiver.Name != "" {
		label = receiver.Name
	}
	debit, credit, err := h.walletUC.Transfer(userID, receiver.ID, label, req.Amount, req.PIN)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "insufficient balance" {
			status = http.StatusPaymentRequired
		} else if err.Error() == "invalid pin" || err.Error() == "pin setup required" || err.Error() == "pin is required" {
			status = http.StatusUnauthorized
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "transfer successful", "transaction": debit, "receiver_transaction": credit})
}
