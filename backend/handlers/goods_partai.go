package handlers

import (
	"food-api/database"
	"food-api/middlewares"
	"food-api/models"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// bungkusOptions = pilihan Bungkus, samakan dengan dropdown di frontend.
var bungkusOptions = []string{"Tikar", "Kain", "Glansing", "Karung Goni", "Karung Plastik", "Yute Rs"}
 
type GoodsPartaiCoaItem struct {
	CoaGoodsID uint  `json:"coa_goods_id"`
	CoaID      *uint `json:"coa_id"`
}
 
type GoodsPartaiInput struct {
	Code    int                  `json:"goods_partai_code"`
	Year    int                  `json:"goods_partai_year"`
	Name    string               `json:"goods_partai_name"`
	Bungkus string               `json:"goods_partai_bungkus"`
	Desc    string               `json:"goods_partai_desc"`
	Active  bool                 `json:"goods_partai_active"`
	Coas    []GoodsPartaiCoaItem `json:"coas"`
}
 
type GoodsPartaiRow struct {
	ID      uint   `json:"goods_partai_id"`
	Code    int    `json:"goods_partai_code"`
	Year    int    `json:"goods_partai_year"`
	Name    string `json:"goods_partai_name"`
	Bungkus string `json:"goods_partai_bungkus"`
	Desc    string `json:"goods_partai_desc"`
	Active  bool   `json:"goods_partai_active"`
}
 
type GoodsPartaiCoaRow struct {
	CoaGoodsID   uint   `json:"coa_goods_id"`
	CoaGoodsName string `json:"coa_goods_name"`
	CoaID        *uint  `json:"coa_id"`
	CoaNo        string `json:"coa_no"`
	CoaName      string `json:"coa_name"`
}
 
type GoodsPartaiDetail struct {
	GoodsPartaiRow
	Coas []GoodsPartaiCoaRow `json:"coas"`
}

// Struct Terra
type GoodsPartaiTarraInput struct {
    BeratMin *float64 `json:"goods_partai_tarra_berat_min"`
    BeratMax *float64 `json:"goods_partai_tarra_berat_max"`
    Nominal  *float64 `json:"goods_partai_tarra_nominal"`
}

type GoodsPartaiDecimalInput struct {
	DecimalMin   *float64 `json:"goods_partai_decimal_min"`
	DecimalMax   *float64 `json:"goods_partai_decimal_max"`
	BrutoNominal *float64 `json:"goods_partai_bruto_nominal"`
}

func activeToInt(active bool) int {
	if active {
		return 1
	}
	return 0
}

func toGoodsPartaiRow(p models.GoodsPartai) GoodsPartaiRow {
	return GoodsPartaiRow{
		ID: p.ID,
		Code: p.Code,
		Year: p.Year,
		Name: p.Name,
		Bungkus: p.Bungkus,
		Desc: p.Desc,
		Active: p.Active == 1,
	}
}

func loadGoodsPartaiCoas(partaiID uint) ([]GoodsPartaiCoaRow, error) {
	rows := []GoodsPartaiCoaRow{}

	err := database.DB.Table("coa_goods AS g").
		Select(`g.coa_goods_id AS coa_goods_id,
			g.coa_goods_name AS coa_goods_name,
			p.coa_id AS coa_id,
			COALESCE(c.coa_no, '') AS coa_no,
			COALESCE(c.coa_name, '') AS coa_name`).
		Joins("LEFT JOIN goods_coa_partai AS p ON p.coa_goods_id = g.coa_goods_id AND p.goods_partai_id = ?", partaiID).
		Joins("LEFT JOIN coas AS c ON c.coa_id = p.coa_id").
		Order("g.coa_goods_id").
		Scan(&rows).Error
 
	return rows, err
}

func loadGoodsPartaiDetail(p models.GoodsPartai) (GoodsPartaiDetail, error) {
	coas, err := loadGoodsPartaiCoas(p.ID)
	return GoodsPartaiDetail{
		GoodsPartaiRow: toGoodsPartaiRow(p),
		Coas: coas,
	}, err
}



func findGoodsPartai(c *gin.Context, userID uint) (models.GoodsPartai, bool) {
	var p models.GoodsPartai

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err == nil {
		err = database.DB.Where("goods_partai_id = ? AND user_id = ?", id, userID).First(&p).Error
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message" : "Partai tidak ditemukan",
		})
		return p, false
	}

	return p, true
}


func validasiGoodsPartai(in *GoodsPartaiInput, userID, id uint) map[string]string {
	errs := map[string]string{}
 
	in.Name = strings.TrimSpace(in.Name)
	in.Desc = strings.TrimSpace(in.Desc)
	in.Bungkus = strings.TrimSpace(in.Bungkus)
 
	if in.Code <= 0 {
		errs["goods_partai_code"] = "Kode partai wajib diisi"
	}
 
	if in.Year < 1990 || in.Year > 2100 {
		errs["goods_partai_year"] = "Tahun tidak valid"
	}
 
	if in.Name == "" {
		errs["goods_partai_name"] = "Nama partai wajib diisi"
	} else if len([]rune(in.Name)) > 256 {
		errs["goods_partai_name"] = "Nama partai maksimal 256 karakter"
	}
 
	if !slices.Contains(bungkusOptions, in.Bungkus) {
		errs["goods_partai_bungkus"] = "Bungkus tidak valid"
	}
 
	if errs["goods_partai_code"] == "" && errs["goods_partai_year"] == "" {
		var n int64
		database.DB.Model(&models.GoodsPartai{}).
			Where("user_id = ? AND goods_partai_code = ? AND goods_partai_year = ? AND goods_partai_id <> ?",
				userID, in.Code, in.Year, id).
			Count(&n)
		if n > 0 {
			errs["goods_partai_code"] = "Kode partai sudah dipakai di tahun ini"
		}
	}
 
	seen := map[uint]bool{}
	typeIDs := []uint{}
	coaIDs := []uint{}
 
	for i := range in.Coas {
		it := &in.Coas[i]
		if it.CoaID != nil && *it.CoaID == 0 {
			it.CoaID = nil
		}
 
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

func syncGoodsPartaiCoas(tx *gorm.DB, partaiID, userID uint, items []GoodsPartaiCoaItem) error {
	var typeIDs []uint
	if err := tx.Model(&models.CoaGoods{}).Order("coa_goods_id").Pluck("coa_goods_id", &typeIDs).Error; err != nil {
		return err
	}
 
	provided := make(map[uint]*uint, len(items))
	for _, it := range items {
		provided[it.CoaGoodsID] = it.CoaID
	}
 
	set := []models.GoodsCoaPartai{}
	missing := []models.GoodsCoaPartai{}
 
	for _, typeID := range typeIDs {
		row := models.GoodsCoaPartai{
			IDGoodsPartai: partaiID, 
			IDCoaGoods: typeID, 
			UserID: userID,
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
			Columns:   []clause.Column{{Name: "goods_partai_id"}, {Name: "coa_goods_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"coa_id", "updated_at"}),
		}).Create(&set).Error
		if err != nil {
			return err
		}
	}
 
	return nil
}

func GetCoaGoods(c *gin.Context) {
	types := []models.CoaGoods{}
 
	if err := database.DB.Order("coa_goods_id").Find(&types).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal ambil data jenis akun",
		})
		return
	}
 
	c.JSON(http.StatusOK, gin.H{
		"data": types,
	})
}


func GetGoodsPartai(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message" : "User tidak dikenali",
		})
		return
	}

	q := database.DB.Where("user_id = ?", userID)

	if s := strings.TrimSpace(c.Query("q")); s != "" {
		like := "%" + strings.ToLower(s) + "%"
		q = q.Where("(LOWER(goods_partai_name) LIKE ? OR CAST(goods_partai_code AS TEXT) LIKE ?)", like, like)
	}

	switch c.Query("active") {
	case "true":
		q = q.Where("goods_partai_active = 1")
	case "false":
		q = q.Where("goods_partai_active <> 1")
	}

	list := []models.GoodsPartai{}
	if err := q.Order("goods_partai_year DESC, goods_partai_id DESC").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message" : "Gagal ambil data partai",
		})
		return
	}

	rows := make([]GoodsPartaiRow, 0, len(list))
	for _, p := range list {
		rows = append(rows, toGoodsPartaiRow(p))
	}

	c.JSON(http.StatusOK, gin.H{
		"data" : rows,
	})
}

func GetGoodsPartaiBy(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message" : "User tidak dikenali",
		})
		return
	}

	partai, found := findGoodsPartai(c, userID)
	if !found {
		return
	}

	detail, err := loadGoodsPartaiDetail(partai)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message" : "Gagal ambil data partai",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data" : detail,
	})
}

func CreateGoodsPartai(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message" : "User tidak dikenali",
		})
		return
	}

	var in GoodsPartaiInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message" : "JSON tidak valid",
		})
		return
	}

	if errs := validasiGoodsPartai(&in, userID, 0); len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors" : errs,
		})
		return
	}

	partai := models.GoodsPartai{
		UserID: userID,
		Code: in.Code,
		Year: in.Year,
		Name: in.Name,
		Bungkus: in.Bungkus,
		Desc: in.Desc,
		Active: activeToInt(in.Active),
	}


	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&partai).Error; err != nil {
			return err
		}
		return syncGoodsPartaiCoas(tx, partai.ID, userID, in.Coas)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message" : "Gagal menyimpan partai",
		})
		return
	}

	detail, err := loadGoodsPartaiDetail(partai)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message" : "Tersimpan, tapi gagal memuat ulang data",
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"data": detail,
	})
}

func UpdateGoodsPartai(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message" : "User tidak dikenali",
		})
		return
	}

	partai, found := findGoodsPartai(c, userID)
	if !found {
		return
	}

	var in GoodsPartaiInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message" : "JSON tidak valid",
		})
		return
	}

	if errs := validasiGoodsPartai(&in, userID, partai.ID); len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors" : errs,
		})
		return
	}

	partai.Code = in.Code
	partai.Year = in.Year
	partai.Name = in.Name
	partai.Bungkus = in.Bungkus
	partai.Desc = in.Desc
	partai.Active = activeToInt(in.Active)


	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&partai).Error; err != nil {
			return err
		}
		return syncGoodsPartaiCoas(tx, partai.ID, userID, in.Coas)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message" : "Gagal menyimpan partai",
		})
		return
	}

	detail, err := loadGoodsPartaiDetail(partai)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message" : "Tersimpan, tapi gagal memuat ulang data",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": detail,
	})
}

func DeleteGoodsPartai(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}
 
	partai, found := findGoodsPartai(c, userID)
	if !found {
		return
	}
 
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("goods_partai_id = ?", partai.ID).Delete(&models.GoodsCoaPartai{}).Error; err != nil {
			return err
		}
		return tx.Delete(&partai).Error
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menghapus partai",
		})
		return
	}
 
	c.Status(http.StatusNoContent)
}

func validasiGoodsPartaiTarra(in *GoodsPartaiTarraInput, partaiID uint) map[string]string {
	errs := map[string]string{}
 
	if in.BeratMin == nil {
		errs["goods_partai_tarra_berat_min"] = "Berat min wajib diisi"
	} else if *in.BeratMin < 0 {
		errs["goods_partai_tarra_berat_min"] = "Berat min tidak boleh negatif"
	}
 
	if in.BeratMax == nil {
		errs["goods_partai_tarra_berat_max"] = "Berat max wajib diisi"
	} else if *in.BeratMax < 0 {
		errs["goods_partai_tarra_berat_max"] = "Berat max tidak boleh negatif"
	}
 
	if in.Nominal == nil {
		errs["goods_partai_tarra_nominal"] = "Nominal wajib diisi"
	} else if *in.Nominal < 0 {
		errs["goods_partai_tarra_nominal"] = "Nominal tidak boleh negatif"
	}
 
	if len(errs) > 0 {
		return errs
	}
 
	if *in.BeratMin > *in.BeratMax {
		errs["goods_partai_tarra_berat_max"] = "Berat max tidak boleh lebih kecil dari berat min"
		return errs
	}
 
	var n int64
	database.DB.Model(&models.GoodsPartaiTarra{}).
		Where("goods_partai_id = ? AND goods_partai_tarra_berat_min <= ? AND goods_partai_tarra_berat_max >= ?",
			partaiID, *in.BeratMax, *in.BeratMin).
		Count(&n)
	if n > 0 {
		errs["goods_partai_tarra_berat_min"] = "Rentang bertabrakan dengan data yang sudah ada"
	}
 
	return errs
}
 
func validasiGoodsPartaiDecimal(in *GoodsPartaiDecimalInput, partaiID uint) map[string]string {
	errs := map[string]string{}
 
	if in.DecimalMin == nil {
		errs["goods_partai_decimal_min"] = "Desimal min wajib diisi"
	} else if *in.DecimalMin < 0 {
		errs["goods_partai_decimal_min"] = "Desimal min tidak boleh negatif"
	}
 
	if in.DecimalMax == nil {
		errs["goods_partai_decimal_max"] = "Desimal max wajib diisi"
	} else if *in.DecimalMax < 0 {
		errs["goods_partai_decimal_max"] = "Desimal max tidak boleh negatif"
	}
 
	if in.BrutoNominal == nil {
		errs["goods_partai_bruto_nominal"] = "Nominal potongan wajib diisi"
	}
 
	if len(errs) > 0 {
		return errs
	}
 
	if *in.DecimalMin > *in.DecimalMax {
		errs["goods_partai_decimal_max"] = "Desimal max tidak boleh lebih kecil dari desimal min"
		return errs
	}
 
	var n int64
	database.DB.Model(&models.GoodsPartaiDecimal{}).
		Where("goods_partai_id = ? AND goods_partai_decimal_min <= ? AND goods_partai_decimal_max >= ?",
			partaiID, *in.DecimalMax, *in.DecimalMin).
		Count(&n)
	if n > 0 {
		errs["goods_partai_decimal_min"] = "Rentang bertabrakan dengan data yang sudah ada"
	}
 
	return errs
}

func GetGoodsPartaiTarra(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message" : "User tidak dikenali",
		})
		return
	}

	partai, found := findGoodsPartai(c, userID)
	if !found {
		return
	}

	list := []models.GoodsPartaiTarra{}
	if err := database.DB.Where("goods_partai_id = ?", partai.ID).Order("goods_partai_tarra_berat_min ASC, goods_partai_tarra_id ASC").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message" : "Gagal ambil data terra",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data" : list,
	})
}

func GetGoodsPartaiDecimal(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message" : "User tidak dikenali",
		})
		return
	}

	partai, found := findGoodsPartai(c, userID)
	if !found {
		return
	}

	list := []models.GoodsPartaiDecimal{}
	if err := database.DB.Where("goods_partai_id = ?", partai.ID).Order("goods_partai_decimal_min ASC, goods_partai_decimal_id ASC").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message" : "Gagal ambil data decimal",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data" : list,
	})
}

func CreateGoodsPartaiTarra(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}
 
	partai, found := findGoodsPartai(c, userID)
	if !found {
		return
	}

	var in GoodsPartaiTarraInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message" : "JSON tidak valid",
		})
		return
	}

	if errs := validasiGoodsPartaiTarra(&in, partai.ID); len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors" : errs,
		})
		return
	}

	row := models.GoodsPartaiTarra{
		IDGoodsPartai: partai.ID,
		BeratMin: *in.BeratMin,
		BeratMax: *in.BeratMax,
		Nominal: *in.Nominal,
	}

	if err := database.DB.Create(&row).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message" : "Gagal menyimpan data",
		})
		return
	}

	c.JSON(http.StatusCreated, row)
}

func CreateGoodsPartaiDecimal(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}
 
	partai, found := findGoodsPartai(c, userID)
	if !found {
		return
	}

	var in GoodsPartaiDecimalInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message" : "JSON tidak valid",
		})
		return
	}

	if errs := validasiGoodsPartaiDecimal(&in, partai.ID); len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors" : errs,
		})
		return
	}

	row := models.GoodsPartaiDecimal{
		IDGoodsPartai: partai.ID,
		DecimalMin: *in.DecimalMin,
		DecimalMax: *in.DecimalMax,
		BrutoNominal: *in.BrutoNominal,
	}

	if err := database.DB.Create(&row).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message" : "Gagal menyimpan data",
		})
		return
	}

	c.JSON(http.StatusCreated, row)
}

func DeleteGoodsPartaiTarra(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}

	partai, found := findGoodsPartai(c, userID)
	if !found {
		return
	}

	tarraID, err := strconv.ParseUint(c.Param("tarra_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Data tarra tidak ditemukan",
		})
		return
	}

	result := database.DB.
		Where("goods_partai_tarra_id = ? AND goods_partai_id = ?", tarraID, partai.ID).
		Delete(&models.GoodsPartaiTarra{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menghapus data tarra",
		})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Data tarra tidak ditemukan",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

func DeleteGoodsPartaiDecimal(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}

	partai, found := findGoodsPartai(c, userID)
	if !found {
		return
	}

	decimalID, err := strconv.ParseUint(c.Param("decimal_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Data tarra tidak ditemukan",
		})
		return
	}

	result := database.DB.
		Where("goods_partai_decimal_id = ? AND goods_partai_id = ?", decimalID, partai.ID).
		Delete(&models.GoodsPartaiDecimal{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menghapus data decimal",
		})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Data decimal tidak ditemukan",
		})
		return
	}

	c.Status(http.StatusNoContent)
}