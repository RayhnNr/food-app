package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"food-api/database"
	"food-api/models"

	"github.com/gin-gonic/gin"
)

type ProductInput struct {
	Kode  string `json:"kode"`
	Nama string `json:"nama"`
	Harga int `json:"harga"`
	IsReady *bool `json:"is_ready"`
	Gambar string `json:"gambar"`
}

func generateKode() (string, error){
	var angka int
	err := database.DB.Model(&models.Product{}).
		Select("COALESCE(MAX(CAST(SUBSTRING(kode FROM 3) AS INTEGER)), 0)").
		Where("kode ~ ?", "^K-[0-9]+$").
		Scan(&angka).Error
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("K-%02d", angka+1), nil
}

func validasi(in *ProductInput) map[string]string {
	errs := map[string]string{}

	// in.Kode = strings.TrimSpace(in.Kode)
	in.Nama = strings.TrimSpace(in.Nama)

	if in.Nama == "" {
		errs["nama"] = "Nama Wajib Diisi!"
	} else if len([]rune(in.Nama)) < 3 {
		errs["nama"] = "Nama harus lebih dari 3 karakter"
	}

	if in.Harga <= 0 {
		errs["harga"] = "Harga tidak boleh 0"
	}

	return errs
}

func GetData(c *gin.Context) {
	products := []models.Product{}

	q := strings.TrimSpace(c.Query("q"))
	db := database.DB.Order("id DESC")
	if q != "" {
		db = db.Where("nama ILIKE ?", "%"+q+"%")
	}

	if err := db.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal ambil data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": products,
	})
}

func CreateProduct(c *gin.Context) {
	var in ProductInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Json tidak valid",
		})
		return
	}

	if errs:= validasi(&in); len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": errs,
		})
		return
	}

	kode, err := generateKode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal membuat kode",
		})
		return
	}

	product := models.Product{
		Kode: kode,
		Nama: in.Nama,
		Harga: in.Harga,
		IsReady: true,
		Gambar: in.Gambar,
	}

	if in.IsReady != nil {
		product.IsReady = *in.IsReady
	}

	if err := database.DB.Create(&product).Error; err != nil{
		log.Printf("gagal menyimpan produk: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menyimpan",
		})
		return
	}
	
	c.JSON(http.StatusCreated, product)

}

func GetProduct(c *gin.Context){
	var product models.Product

	if err := database.DB.First(&product, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"Message": "Produk tidak ditemukan",
		})
		return
	}
	c.JSON(http.StatusOK, product)
}

func UpdateProduct(c *gin.Context){
	var product models.Product

	if err := database.DB.First(&product, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"Message": "Produk tidak ditemukan",
		})
		return
	}

	var in ProductInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Json tidak valid",
		})
		return
	}

	if errs:= validasi(&in); len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": errs,
		})
		return
	}

	product.Nama = in.Nama
	product.Harga = in.Harga
	product.Gambar = in.Gambar
	if in.IsReady != nil {
		product.IsReady = *in.IsReady
	}

	if err := database.DB.Save(&product).Error; err != nil{
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal memperbarui",
		})
		return
	}
	c.JSON(http.StatusOK, product)
}

func DeleteProduct(c *gin.Context) {
	result := database.DB.Delete(&models.Product{}, c.Param("id"))
	
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menghapus",
		})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Product tidak dapat ditemukan",
		})
		return
	}
	c.Status(http.StatusNoContent)
}

