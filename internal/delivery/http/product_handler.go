package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetProducts(c *gin.Context) {
	products, err := h.productUC.GetAll()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, products)
}

func (h *Handler) GetDetail(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	product, err := h.productUC.GetByID(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}

	c.JSON(200, product)
}
