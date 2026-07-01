package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Checkout(c *gin.Context) {
	userID := c.GetInt64("user_id")

	order, err := h.orderUC.Checkout(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	intent, err := h.walletUC.CreatePaymentIntent(order.ID, userID, order.TotalPrice, "Fragrance Store")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "order created but failed to create payment intent: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"order": order,
		"payment_intent": gin.H{
			"token":         intent.Token,
			"amount":        intent.Amount,
			"merchant":      intent.MerchantName,
			"merchant_name": intent.MerchantName,
			"status":        intent.Status,
			"deep_link":     intent.DeepLink,
			"expires_at":    intent.ExpiresAt,
			"reference":     "Order #" + strconv.FormatInt(order.ID, 10),
			"note":          "Pembayaran e-commerce melalui Kantongin",
		},
	})
}

func (h *Handler) GetMyOrders(c *gin.Context) {
	userID := c.GetInt64("user_id")

	orders, err := h.orderUC.GetMyOrders(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (h *Handler) GetOrderDetail(c *gin.Context) {
	userID := c.GetInt64("user_id")
	orderID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	order, items, err := h.orderUC.GetOrderDetail(orderID, userID)
	if err != nil {
		if err.Error() == "forbidden" {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order": order,
		"items": items,
	})
}

func (h *Handler) CreateOrderPaymentIntent(c *gin.Context) {
	userID := c.GetInt64("user_id")
	orderID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	order, _, err := h.orderUC.GetOrderDetail(orderID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if order.PaymentStatus == "paid" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order already paid"})
		return
	}

	intent, err := h.walletUC.CreatePaymentIntent(order.ID, userID, order.TotalPrice, "Fragrance Store")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":         intent.Token,
		"amount":        intent.Amount,
		"merchant":      intent.MerchantName,
		"merchant_name": intent.MerchantName,
		"status":        intent.Status,
		"deep_link":     intent.DeepLink,
		"expires_at":    intent.ExpiresAt,
		"reference":     "Order #" + strconv.FormatInt(order.ID, 10),
		"note":          "Pembayaran e-commerce melalui Kantongin",
	})
}
