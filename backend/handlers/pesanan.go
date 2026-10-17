package handlers

import (
	"net/http"
	"strings"

	"food-api/database"
	"food-api/models"

	"github.com/gin-gonic/gin"
)

// type PesananInput struct {
// 	ProductID  uint   `json:"product_id"`
// 	Jumlah     int    `json:"jumlah"`
// 	Keterangan string `json:"keterangan"`
// }

func GetPesanans(c *gin.Context) {
	pesanans := []models.Pesanan{}

	db := database.DB.Preload("Items").Order("id DESC")

	if status := strings.TrimSpace(c.Query("status")); status != "" {
		db = db.Where("status = ?", status)
	}
	if q := strings.TrimSpace(c.Query("q")); q != "" {
		db = db.Where("nama ILIKE ?", "%"+q+"%")
	}

	if err := db.Find(&pesanans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal ambil data pesanan"})
		return
	}

	c.JSON(http.StatusOK, pesanans)
}

func GetPesanan(c *gin.Context) {
	var pesanan models.Pesanan

	if err := database.DB.Preload("Items").First(&pesanan, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Pesanan tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, pesanan)
}

