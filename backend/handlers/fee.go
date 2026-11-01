package handlers

import (
	"food-api/database"
	"food-api/middlewares"
	"food-api/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var feeDefaults = []struct {
	Slug string
	Name string
}{
	{"ganti_tikar", "Ganti Tikar"},
	{"ganti_tali", "Ganti Tali"},
	{"ongkos_kuli", "Ongkos Kuli"},
	{"pajak", "Pajak"},
}

type FeeItem struct {
	Slug string   `json:"fee_slug"`
	Cost *float64 `json:"fee_cost"`
}

type FeeInput struct {
	Fees []FeeItem `json:"fees"`
}

type FeeRow struct {
	Slug string  `json:"fee_slug"`
	Name string  `json:"fee_name"`
	Cost float64 `json:"fee_cost"`
}

// ensureFees membuat biaya bawaan yang belum ada untuk user, nominalnya 0.
func ensureFees(tx *gorm.DB, userID uint) error {
	rows := make([]models.Fee, 0, len(feeDefaults))
	for _, d := range feeDefaults {
		rows = append(rows, models.Fee{UserID: userID, Slug: d.Slug, Name: d.Name})
	}

	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "fee_slug"}},
		DoNothing: true,
	}).Create(&rows).Error
}

func loadFees(tx *gorm.DB, userID uint) ([]FeeRow, error) {
	list := []models.Fee{}
	if err := tx.Where("user_id = ?", userID).Order("fee_id ASC").Find(&list).Error; err != nil {
		return nil, err
	}

	rows := make([]FeeRow, 0, len(list))
	for _, f := range list {
		rows = append(rows, FeeRow{Slug: f.Slug, Name: f.Name, Cost: f.Cost})
	}

	return rows, nil
}

func validasiFee(in *FeeInput) map[string]string {
	errs := map[string]string{}

	if len(in.Fees) == 0 {
		errs["fees"] = "Data biaya wajib diisi"
		return errs
	}

	known := map[string]bool{}
	for _, d := range feeDefaults {
		known[d.Slug] = true
	}

	seen := map[string]bool{}

	for i := range in.Fees {
		it := &in.Fees[i]
		it.Slug = strings.ToLower(strings.TrimSpace(it.Slug))
		key := "fees." + it.Slug

		switch {
		case !known[it.Slug]:
			errs["fees"] = "Jenis biaya tidak dikenal: " + it.Slug
		case seen[it.Slug]:
			errs[key] = "Biaya dikirim lebih dari satu kali"
		case it.Cost == nil:
			errs[key] = "Nominal wajib diisi"
		case *it.Cost < 0:
			errs[key] = "Nominal tidak boleh negatif"
		}

		seen[it.Slug] = true
	}

	return errs
}

// GET /fee
func GetFee(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}

	if err := ensureFees(database.DB, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menyiapkan data biaya",
		})
		return
	}

	rows, err := loadFees(database.DB, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal ambil data biaya",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": rows,
	})
}

// PUT /fee
func UpdateFee(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}

	var in FeeInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "JSON tidak valid",
		})
		return
	}

	if errs := validasiFee(&in); len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": errs,
		})
		return
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := ensureFees(tx, userID); err != nil {
			return err
		}

		for _, it := range in.Fees {
			err := tx.Model(&models.Fee{}).
				Where("user_id = ? AND fee_slug = ?", userID, it.Slug).
				Update("fee_cost", *it.Cost).Error
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menyimpan biaya",
		})
		return
	}

	rows, err := loadFees(database.DB, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Tersimpan, tapi gagal memuat ulang data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": rows,
	})
}