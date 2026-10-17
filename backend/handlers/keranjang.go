package handlers

import (
	"net/http"
	"strings"
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


func TambahKeranjang(c *gin.Context) {
	var in KeranjangInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message" : "Json tidak valid",
		})
		return
	}

	errs := map[string]string{}

	if in.ProductID  == 0 {
		errs["product_id"] = "Product Wajib dipilih"
	}
	if in.Jumlah <= 0 {
		errs["jumlah"] = "Jumlah minimal 1"
	}
	if len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors" : errs,
		})
		return
	}

	var product models.Product
	if err := database.DB.First(&product, in.ProductID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message" : "Produk tidak ditemukan",
		})
		return
	}
	if !product.IsReady {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors" : map[string]string{"product_id": "Produk sedang habis"},
		})
		return
	}

	keterangan := strings.TrimSpace(in.Keterangan)
	var item models.Keranjang
	err := database.DB.Where("product_id = ?", in.ProductID).First(&item).Error
	if err == nil {
		item.Jumlah += in.Jumlah
		if keterangan != "" {
			item.Keterangan = keterangan
		}
		database.DB.Save(&item)
	} else {
		item = models.Keranjang{
			ProductID: in.ProductID,
			Jumlah: in.Jumlah,
			Keterangan: keterangan,
		}
		if err := database.DB.Create(&item).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message" : "Gagal menyimpan",
			})
			return
		}

	}

	database.DB.Preload("Product").First(&item, item.ID)
	c.JSON(http.StatusCreated, item)
}

func UpdateKeranjang(c *gin.Context){
	var item models.Keranjang
	if err := database.DB.First(&item, c.Param("id")).Error; err != nil{
		c.JSON(http.StatusNotFound, gin.H{
			"mesaage" : "Produk item tidak ditemukan",
		})
	}

	var in KeranjangInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message" : "Json tidak valid",
		})
	}

	if in.Jumlah <= 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors" : map[string]string{
				"jumlah" : "jumlah minimal 1",
			},
		})
		return
	}

	item.Jumlah = in.Jumlah
	item.Keterangan = strings.TrimSpace(in.Keterangan)
	database.DB.Save(&item)

	database.DB.Preload("Product").First(&item, item.ID)
	c.JSON(http.StatusOK, item)

}

func DeleteKeranjang(c *gin.Context) {
	result := database.DB.Delete(&models.Keranjang{}, c.Param("id"))
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menghapus"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Item tidak ditemukan"})
		return
	}
	c.Status(http.StatusNoContent)
}

func ClearKeranjang(c *gin.Context) {
	if err := database.DB.Where("1 = 1").Delete(&models.Keranjang{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal mengosongkan"})
		return
	}
	c.Status(http.StatusNoContent)
}