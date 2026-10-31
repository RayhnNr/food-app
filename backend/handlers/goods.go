package handlers

import (
	"food-api/database"
	"food-api/middlewares"
	"food-api/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GoodsCoaItem struct {
	CoaGoodsID uint  `json:"coa_goods_id"`
	CoaID      *uint `json:"coa_id"`
}

type GoodsInput struct {
	Name       string         `json:"goods_name"`
	Desc       string         `json:"goods_desc"`
	Active     bool           `json:"goods_active"`
	CategoryID *uint          `json:"goods_category_id"`
	BrandID    *uint          `json:"goods_brand_id"`
	UnitID     *uint          `json:"goods_unit_id"`
	Coas       []GoodsCoaItem `json:"coas"`
}

type GoodsRow struct {
	ID           uint   `json:"goods_id"`
	Name         string `json:"goods_name"`
	Desc         string `json:"goods_desc"`
	Active       bool   `json:"goods_active"`
	CategoryID   *uint  `json:"goods_category_id"`
	CategoryName string `json:"goods_category_name"`
	BrandID      *uint  `json:"goods_brand_id"`
	UnitID       *uint  `json:"goods_unit_id"`
}

type GoodsCoaRow struct {
	CoaGoodsID   uint   `json:"coa_goods_id"`
	CoaGoodsName string `json:"coa_goods_name"`
	CoaID        *uint  `json:"coa_id"`
	CoaNo        string `json:"coa_no"`
	CoaName      string `json:"coa_name"`
}

type GoodsDetail struct {
	GoodsRow
	Coas []GoodsCoaRow `json:"coas"`
}

type GoodsCategoryOption struct {
	ID   uint   `json:"goods_category_id" gorm:"column:goods_category_id"`
	Name string `json:"goods_category_name" gorm:"column:goods_category_name"`
}

type GoodsCategoryInput struct {
	Name string `json:"goods_category_name"`
}

func toGoodsRow(g models.Goods) GoodsRow {
	row := GoodsRow{
		ID:         g.ID,
		Name:       g.Name,
		Desc:       g.Desc,
		Active:     g.Active == 1,
		CategoryID: g.IDGoodsCategory,
		BrandID:    g.BrandID,
		UnitID:     g.UnitID,
	}

	if g.GoodsCategory != nil {
		row.CategoryName = g.GoodsCategory.Name
	}

	return row
}

func loadGoodsCoas(goodsID uint) ([]GoodsCoaRow, error) {
	rows := []GoodsCoaRow{}

	err := database.DB.Table("coa_goods AS g").
		Select(`g.coa_goods_id AS coa_goods_id,
			g.coa_goods_name AS coa_goods_name,
			p.coa_id AS coa_id,
			COALESCE(c.coa_no, '') AS coa_no,
			COALESCE(c.coa_name, '') AS coa_name`).
		Joins("LEFT JOIN goods_coa AS p ON p.coa_goods_id = g.coa_goods_id AND p.goods_id = ?", goodsID).
		Joins("LEFT JOIN coas AS c ON c.coa_id = p.coa_id").
		Order("g.coa_goods_id").
		Scan(&rows).Error

	return rows, err
}

func loadGoodsDetail(goodsID uint) (GoodsDetail, error) {
	var g models.Goods
	if err := database.DB.Preload("GoodsCategory").First(&g, goodsID).Error; err != nil {
		return GoodsDetail{}, err
	}

	coas, err := loadGoodsCoas(g.ID)
	return GoodsDetail{
		GoodsRow: toGoodsRow(g),
		Coas:     coas,
	}, err
}

func findGoods(c *gin.Context, userID uint) (models.Goods, bool) {
	var g models.Goods

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err == nil {
		err = database.DB.Where("goods_id = ? AND user_id = ?", id, userID).First(&g).Error
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Barang tidak ditemukan",
		})
		return g, false
	}

	return g, true
}

func zeroToNil(v *uint) *uint {
	if v != nil && *v == 0 {
		return nil
	}
	return v
}

func validasiGoods(in *GoodsInput, userID, id uint) map[string]string {
	errs := map[string]string{}

	in.Name = strings.TrimSpace(in.Name)
	in.Desc = strings.TrimSpace(in.Desc)
	in.CategoryID = zeroToNil(in.CategoryID)
	in.BrandID = zeroToNil(in.BrandID)
	in.UnitID = zeroToNil(in.UnitID)

	if in.Name == "" {
		errs["goods_name"] = "Nama barang wajib diisi"
	} else if len([]rune(in.Name)) > 256 {
		errs["goods_name"] = "Nama barang maksimal 256 karakter"
	} else {
		var n int64
		database.DB.Model(&models.Goods{}).
			Where("user_id = ? AND LOWER(goods_name) = LOWER(?) AND goods_id <> ?", userID, in.Name, id).
			Count(&n)
		if n > 0 {
			errs["goods_name"] = "Nama barang sudah dipakai"
		}
	}

	if in.CategoryID != nil {
		var n int64
		database.DB.Model(&models.GoodsCategory{}).
			Where("goods_category_id = ? AND user_id = ?", *in.CategoryID, userID).
			Count(&n)
		if n == 0 {
			errs["goods_category_id"] = "Kategori tidak ditemukan"
		}
	}

	seen := map[uint]bool{}
	typeIDs := []uint{}
	coaIDs := []uint{}

	for i := range in.Coas {
		it := &in.Coas[i]
		it.CoaID = zeroToNil(it.CoaID)

		key := "coas." + strconv.FormatUint(uint64(it.CoaGoodsID), 10)
		if seen[it.CoaGoodsID] {
			errs[key] = "Jenis akun dikirim lebih dari satu kali"
			continue
		}
		seen[it.CoaGoodsID] = true

		typeIDs = append(typeIDs, it.CoaGoodsID)
		if it.CoaID != nil {
			coaIDs = append(coaIDs, *it.CoaID)
		}
	}

	validTypes := map[uint]bool{}
	if len(typeIDs) > 0 {
		var types []models.CoaGoods
		database.DB.Select("coa_goods_id").Where("coa_goods_id IN ?", typeIDs).Find(&types)
		for _, t := range types {
			validTypes[t.ID] = true
		}
	}

	validCoas := map[uint]bool{}
	if len(coaIDs) > 0 {
		var coas []models.Coa
		database.DB.Select("coa_id", "coa_no").Where("coa_id IN ?", coaIDs).Find(&coas)
		for _, co := range coas {
			if strings.HasSuffix(strings.TrimSpace(co.CoaNo), ",00") {
				validCoas[co.CoaID] = true
			}
		}
	}

	for _, it := range in.Coas {
		key := "coas." + strconv.FormatUint(uint64(it.CoaGoodsID), 10)
		if _, bad := errs[key]; bad {
			continue
		}

		if !validTypes[it.CoaGoodsID] {
			errs[key] = "Jenis akun tidak ditemukan"
		} else if it.CoaID != nil && !validCoas[*it.CoaID] {
			errs[key] = "Akun tidak ditemukan atau bukan akun detail"
		}
	}

	return errs
}

func syncGoodsCoas(tx *gorm.DB, goodsID, userID uint, items []GoodsCoaItem) error {
	var typeIDs []uint
	if err := tx.Model(&models.CoaGoods{}).Order("coa_goods_id").Pluck("coa_goods_id", &typeIDs).Error; err != nil {
		return err
	}

	provided := make(map[uint]*uint, len(items))
	for _, it := range items {
		provided[it.CoaGoodsID] = it.CoaID
	}

	set := []models.GoodsCoa{}
	missing := []models.GoodsCoa{}

	for _, typeID := range typeIDs {
		row := models.GoodsCoa{
			IDGoods:    goodsID,
			IDCoaGoods: typeID,
			UserID:     userID,
		}

		if coaID, ok := provided[typeID]; ok {
			row.IDCoa = coaID
			set = append(set, row)
		} else {
			missing = append(missing, row)
		}
	}

	if len(missing) > 0 {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&missing).Error; err != nil {
			return err
		}
	}

	if len(set) > 0 {
		err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "goods_id"}, {Name: "coa_goods_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"coa_id"}),
		}).Create(&set).Error
		if err != nil {
			return err
		}
	}

	return nil
}

// GET /goods?q=&active=true|false&category_id=
func GetGoods(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}

	q := database.DB.Preload("GoodsCategory").Where("user_id = ?", userID)

	if s := strings.TrimSpace(c.Query("q")); s != "" {
		q = q.Where("LOWER(goods_name) LIKE ?", "%"+strings.ToLower(s)+"%")
	}

	switch c.Query("active") {
	case "true":
		q = q.Where("goods_active = 1")
	case "false":
		q = q.Where("goods_active <> 1")
	}

	if raw := c.Query("category_id"); raw != "" {
		categoryID, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Kategori tidak valid",
			})
			return
		}
		q = q.Where("goods_category_id = ?", categoryID)
	}

	list := []models.Goods{}
	if err := q.Order("goods_name ASC, goods_id ASC").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal ambil data barang",
		})
		return
	}

	rows := make([]GoodsRow, 0, len(list))
	for _, g := range list {
		rows = append(rows, toGoodsRow(g))
	}

	c.JSON(http.StatusOK, gin.H{
		"data": rows,
	})
}

// GET /goods/:id
func GetGoodsBy(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}

	goods, found := findGoods(c, userID)
	if !found {
		return
	}

	detail, err := loadGoodsDetail(goods.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal ambil data barang",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": detail,
	})
}

// POST /goods
func CreateGoods(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}

	var in GoodsInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "JSON tidak valid",
		})
		return
	}

	if errs := validasiGoods(&in, userID, 0); len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": errs,
		})
		return
	}

	goods := models.Goods{
		UserID:          userID,
		Name:            in.Name,
		Desc:            in.Desc,
		Active:          activeToInt(in.Active),
		IDGoodsCategory: in.CategoryID,
		BrandID:         in.BrandID,
		UnitID:          in.UnitID,
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&goods).Error; err != nil {
			return err
		}
		return syncGoodsCoas(tx, goods.ID, userID, in.Coas)
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menyimpan barang",
		})
		return
	}

	detail, err := loadGoodsDetail(goods.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Tersimpan, tapi gagal memuat ulang data",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": detail,
	})
}

// PUT /goods/:id
func UpdateGoods(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}

	goods, found := findGoods(c, userID)
	if !found {
		return
	}

	var in GoodsInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "JSON tidak valid",
		})
		return
	}

	if errs := validasiGoods(&in, userID, goods.ID); len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": errs,
		})
		return
	}

	goods.Name = in.Name
	goods.Desc = in.Desc
	goods.Active = activeToInt(in.Active)
	goods.IDGoodsCategory = in.CategoryID
	goods.BrandID = in.BrandID
	goods.UnitID = in.UnitID

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&goods).Error; err != nil {
			return err
		}
		return syncGoodsCoas(tx, goods.ID, userID, in.Coas)
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menyimpan barang",
		})
		return
	}

	detail, err := loadGoodsDetail(goods.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Tersimpan, tapi gagal memuat ulang data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": detail,
	})
}

// DELETE /goods/:id
func DeleteGoods(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}

	goods, found := findGoods(c, userID)
	if !found {
		return
	}

	var used int64
	err := database.DB.Model(&models.GoodsPartai{}).
		Where("ref_goods_id = ? AND user_id = ?", goods.ID, userID).
		Count(&used).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menghapus barang",
		})
		return
	}
	if used > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"message": "Barang masih dipakai oleh data partai, tidak bisa dihapus",
		})
		return
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("goods_id = ?", goods.ID).Delete(&models.GoodsCoa{}).Error; err != nil {
			return err
		}
		return tx.Delete(&goods).Error
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menghapus barang",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// GET /goods-category-options
// GET /goods-category-options
func GetGoodsCategoryOptions(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}

	options := []GoodsCategoryOption{}
	err := database.DB.Model(&models.GoodsCategory{}).
		Select("goods_category_id, goods_category_name").
		Where("user_id = ? AND goods_category_active = 1", userID).
		Order("goods_category_name ASC").
		Scan(&options).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal ambil data kategori",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": options,
	})
}

// POST /goods-categories
func CreateGoodsCategory(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}

	var in GoodsCategoryInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "JSON tidak valid",
		})
		return
	}

	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" || len([]rune(in.Name)) > 256 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": map[string]string{
				"goods_category_name": "Nama kategori wajib diisi dan maksimal 256 karakter",
			},
		})
		return
	}

	var count int64
	if err := database.DB.Model(&models.GoodsCategory{}).
		Where("user_id = ? AND LOWER(goods_category_name) = LOWER(?)", userID, in.Name).
		Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal memeriksa kategori",
		})
		return
	}
	if count > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": map[string]string{
				"goods_category_name": "Nama kategori sudah dipakai",
			},
		})
		return
	}

	category := models.GoodsCategory{
		Name:   in.Name,
		Active: 1,
		UserID: userID,
	}
	if err := database.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menyimpan kategori",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": GoodsCategoryOption{
			ID:   category.ID,
			Name: category.Name,
		},
	})
}
