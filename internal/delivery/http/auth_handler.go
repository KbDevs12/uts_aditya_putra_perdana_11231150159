package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Register(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
		Name  string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token and name are required"})
		return
	}

	if err := h.authUC.Register(req.Token, req.Name); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "email not verified" {
			status = http.StatusForbidden
		} else if err.Error() == "user already registered" {
			status = http.StatusConflict
		} else if err.Error() == "invalid firebase token" {
			status = http.StatusUnauthorized
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "registration successful, please login"})
}

func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Token string `json:"token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	jwtToken, err := h.authUC.Login(req.Token)
	if err != nil {
		status := http.StatusUnauthorized
		if err.Error() == "email not verified" {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": jwtToken,
		"token_type":   "Bearer",
	})
}
