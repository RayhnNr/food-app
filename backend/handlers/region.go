package handlers

import (
	"food-api/database"
	"food-api/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RegionOption struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func wilayahAda(table, where string, args ...any) (bool, error) {
	var n int64
	if err := database.DB.Table(table).Where(where, args...).Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

func regionParam(c *gin.Context, notFound string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": notFound,
		})
		return 0, false
	}

	return uint(id), true
}

// GET /provinces
func GetProvinces(c *gin.Context) {
	options := []RegionOption{}
	err := database.DB.Model(&models.Province{}).
		Select("province_id AS id, province AS name").
		Order("province ASC").
		Scan(&options).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal ambil data provinsi",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": options,
	})
}

// GET /provinces/:id/cities
func GetCities(c *gin.Context) {
	provinceID, ok := regionParam(c, "Provinsi tidak ditemukan")
	if !ok {
		return
	}

	found, err := wilayahAda("province", "province_id = ?", provinceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal memeriksa data provinsi",
		})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Provinsi tidak ditemukan",
		})
		return
	}

	options := []RegionOption{}
	err = database.DB.Model(&models.City{}).
		Select("city_id AS id, TRIM(CONCAT(type, ' ', city_name)) AS name").
		Where("province_id = ?", provinceID).
		Order("city_name ASC").
		Scan(&options).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal ambil data kota",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": options,
	})
}

// GET /cities/:id/subdistricts
func GetSubdistricts(c *gin.Context) {
	cityID, ok := regionParam(c, "Kota tidak ditemukan")
	if !ok {
		return
	}

	found, err := wilayahAda("city", "city_id = ?", cityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal memeriksa data kota",
		})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Kota tidak ditemukan",
		})
		return
	}

	options := []RegionOption{}
	err = database.DB.Model(&models.Subdistrict{}).
		Select("subdistrict_id AS id, subdistrict_name AS name").
		Where("city_id = ?", cityID).
		Order("subdistrict_name ASC").
		Scan(&options).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal ambil data kecamatan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": options,
	})
}
