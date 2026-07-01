package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Register(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
		Name  string `json:"name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "firebase token is required"})
		return
	}

	if err := h.authUC.Register(req.Token, req.Name); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "invalid firebase token" {
			status = http.StatusUnauthorized
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "registration successful, otp has been sent to email",
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "firebase token is required"})
		return
	}

	jwtToken, requires2FA, twoFactorMethod, err := h.authUC.Login(req.Token)
	if err != nil {
		status := http.StatusUnauthorized
		if err.Error() == "email not verified" {
			status = http.StatusForbidden
		} else if err.Error() == "user not found, please register first" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":      jwtToken,
		"token_type":        "Bearer",
		"requires_2fa":      requires2FA,
		"two_factor_method": twoFactorMethod,
	})
}

func (h *Handler) SendEmailOTP(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}
	if err := h.authUC.SendEmailOTP(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "otp sent"})
}

func (h *Handler) VerifyEmailOTP(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
		Code  string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and code are required"})
		return
	}
	jwtToken, err := h.authUC.VerifyEmailOTP(req.Email, req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":      "email verified",
		"access_token": jwtToken,
		"token_type":   "Bearer",
	})
}

func (h *Handler) SetupTwoFactor(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req struct {
		Method string `json:"method"`
	}
	_ = c.ShouldBindJSON(&req)
	result, err := h.authUC.SetupTwoFactor(userID, req.Method)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) VerifyTwoFactor(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req struct {
		Method string `json:"method"`
		Code   string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "method and code are required"})
		return
	}
	if err := h.authUC.VerifyTwoFactor(userID, req.Method, req.Code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "2fa verified"})
}

func (h *Handler) SaveFCMToken(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fcm token is required"})
		return
	}
	if err := h.authUC.SaveFCMToken(userID, req.Token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "fcm token saved"})
}
