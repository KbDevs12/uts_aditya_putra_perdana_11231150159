package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type topUpWalletRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
}

func (h *Handler) GetWallet(c *gin.Context) {
	userID := c.GetInt64("user_id")
	wallet, err := h.walletUC.GetWallet(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, wallet)
}

func (h *Handler) TopUpWallet(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req topUpWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount is required and must be greater than 0"})
		return
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

func (h *Handler) PayPaymentIntent(c *gin.Context) {
	userID := c.GetInt64("user_id")
	token := c.Param("token")
	intent, trx, err := h.walletUC.PayPaymentIntent(userID, token)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "insufficient balance" {
			status = http.StatusPaymentRequired
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
