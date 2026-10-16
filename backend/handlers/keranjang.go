package handlers

import (
	"net/http"
	// "strings"

	"food-api/database"
	"food-api/models"

	"github.com/gin-gonic/gin"
)

type KeranjangInput struct {
	ProductID  uint   `json:"product_id"`
	Jumlah     int    `json:"jumlah"`
	Keterangan string `json:"keterangan"`
}

func GetKeranjang(c *gin.Context) {
	items := []models.Keranjang{}
	if err := database.DB.Preload("Product").Order("id DESC").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal ambil data"})
		return
	}
	c.JSON(http.StatusOK, items)
}

