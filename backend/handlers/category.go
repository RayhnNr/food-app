package handlers

import (
	"food-api/database"
	"food-api/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type CategoryInput struct {
	Nama      string `json:"nama"`
	Deskripsi string `json:"deskripsi"`
}

func validasiCategory(in *CategoryInput) map[string]string {
	err := map[string]string{}

	in.Nama = strings.TrimSpace(in.Nama)

	if in.Nama == "" {
		err["nama"] = "Nama kategori wajib diisi"
	} else if len([]rune(in.Nama)) < 3 {
		err["nama"] = "Nama kategori minimal 3 karakter"
	}

	return err
}

func GetCategory(c *gin.Context) {
	category := []models.Category{}

	if err := database.DB.Order("id DESC").Find(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message" : "Gagal ambil data Category",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data" : category,
	})
}


func CreateCategory(c *gin.Context) {
	var in CategoryInput

	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message" : "JSON tidak valid",
		})
		return
	}

	if errs := validasiCategory(&in); len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors" : errs,
		})
		return
	}

	category := models.Category{
		Nama: in.Nama,
		Deskripsi: in.Deskripsi,
	}

	if err := database.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message" : "Gagal menyimpan data",
		})
	}

	c.JSON(http.StatusCreated, category)
}

func GetCategoryBy(c *gin.Context){

	var category models.Category

	if err := database.DB.First(&category, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message" : "Category tidak ditemukan",
		})
		return
	}
	c.JSON(http.StatusOK, category)
}

func UpdateCategory(c *gin.Context) {
	var category models.Category

	if err := database.DB.First(&category, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message" : "Category tidak ditemukan",
		})
		return
	}

	var in CategoryInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message" : "JSON tidak valid",
		})
		return
	}

	if errs := validasiCategory(&in); len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors" : errs,
		})
		return
	}

	category.Nama = in.Nama
	category.Deskripsi = in.Deskripsi

	if err := database.DB.Save(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message" : "Gagal menyimpan data",
		})
	}

	c.JSON(http.StatusCreated, category)
}

func DeleteCategory(c *gin.Context) {
	result := database.DB.Delete(&models.Category{}, c.Param("id"))
	
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menghapus",
		})
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Category tidak dapat ditemukan",
		})
	}
	c.Status(http.StatusNoContent)
}

