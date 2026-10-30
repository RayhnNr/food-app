package handlers

import (
	"food-api/database"
	"food-api/middlewares"
	"food-api/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type GoodsGradeInput struct {
	Name   string   `json:"goods_grade_name"`
	Desc   string   `json:"goods_grade_desc"`
	Price  *float64 `json:"goods_grade_price"`
	Active bool     `json:"goods_grade_active"`
}

type GoodsGradeRow struct {
	ID     uint    `json:"goods_grade_id"`
	Name   string  `json:"goods_grade_name"`
	Desc   string  `json:"goods_grade_desc"`
	Price  float64 `json:"goods_grade_price"`
	Active bool    `json:"goods_grade_active"`
}

func toGoodsGradeRow(g models.GoodsGrade) GoodsGradeRow {
	return GoodsGradeRow{
		ID:     g.ID,
		Name:   g.Name,
		Desc:   g.Desc,
		Price:  g.Price,
		Active: g.Active == 1,
	}
}

func findGoodsGrade(c *gin.Context, userID uint) (models.GoodsGrade, bool) {
	var g models.GoodsGrade

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err == nil {
		err = database.DB.Where("goods_grade_id = ? AND user_id = ?", id, userID).First(&g).Error
	}
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Grade tidak ditemukan"})
		return g, false
	}

	return g, true
}

func validasiGoodsGrade(in *GoodsGradeInput, userID, id uint) map[string]string {
	errs := map[string]string{}

	in.Name = strings.TrimSpace(in.Name)
	in.Desc = strings.TrimSpace(in.Desc)

	if in.Name == "" {
		errs["goods_grade_name"] = "Nama grade wajib diisi"
	} else if len([]rune(in.Name)) > 256 {
		errs["goods_grade_name"] = "Nama grade maksimal 256 karakter"
	} else {
		var count int64
		err := database.DB.Model(&models.GoodsGrade{}).
			Where("user_id = ? AND LOWER(goods_grade_name) = LOWER(?) AND goods_grade_id <> ?", userID, in.Name, id).
			Count(&count).Error
		if err != nil {
			errs["goods_grade_name"] = "Gagal memeriksa nama grade"
		} else if count > 0 {
			errs["goods_grade_name"] = "Nama grade sudah dipakai"
		}
	}

	if in.Price == nil {
		errs["goods_grade_price"] = "Harga wajib diisi"
	} else if *in.Price < 0 {
		errs["goods_grade_price"] = "Harga tidak boleh negatif"
	}

	return errs
}

// GET /goods-grade?q=&active=true|false
func GetGoodsGrade(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User tidak dikenali"})
		return
	}

	q := database.DB.Where("user_id = ?", userID)

	if s := strings.TrimSpace(c.Query("q")); s != "" {
		q = q.Where("LOWER(goods_grade_name) LIKE ?", "%"+strings.ToLower(s)+"%")
	}

	switch c.Query("active") {
	case "true":
		q = q.Where("goods_grade_active = 1")
	case "false":
		q = q.Where("goods_grade_active <> 1")
	}

	list := []models.GoodsGrade{}
	if err := q.Order("goods_grade_price DESC, goods_grade_id ASC").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal ambil data grade"})
		return
	}

	rows := make([]GoodsGradeRow, 0, len(list))
	for _, g := range list {
		rows = append(rows, toGoodsGradeRow(g))
	}

	c.JSON(http.StatusOK, gin.H{"data": rows})
}

// GET /goods-grade/:id
func GetGoodsGradeBy(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User tidak dikenali"})
		return
	}

	grade, found := findGoodsGrade(c, userID)
	if !found {
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": toGoodsGradeRow(grade)})
}

// POST /goods-grade
func CreateGoodsGrade(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User tidak dikenali"})
		return
	}

	var in GoodsGradeInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "JSON tidak valid"})
		return
	}

	if errs := validasiGoodsGrade(&in, userID, 0); len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": errs})
		return
	}

	grade := models.GoodsGrade{
		Name:   in.Name,
		Desc:   in.Desc,
		Price:  *in.Price,
		Active: activeToInt(in.Active),
		UserID: userID,
	}

	if err := database.DB.Create(&grade).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menyimpan grade"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": toGoodsGradeRow(grade)})
}

// PUT /goods-grade/:id
func UpdateGoodsGrade(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User tidak dikenali"})
		return
	}

	grade, found := findGoodsGrade(c, userID)
	if !found {
		return
	}

	var in GoodsGradeInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "JSON tidak valid"})
		return
	}

	if errs := validasiGoodsGrade(&in, userID, grade.ID); len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": errs})
		return
	}

	grade.Name = in.Name
	grade.Desc = in.Desc
	grade.Price = *in.Price
	grade.Active = activeToInt(in.Active)

	if err := database.DB.Save(&grade).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal mengubah grade"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": toGoodsGradeRow(grade)})
}

// DELETE /goods-grade/:id
func DeleteGoodsGrade(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User tidak dikenali"})
		return
	}

	grade, found := findGoodsGrade(c, userID)
	if !found {
		return
	}

	if err := database.DB.Delete(&grade).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menghapus grade"})
		return
	}

	c.Status(http.StatusNoContent)
}