package handlers

import (
	"food-api/database"
	"food-api/middlewares"
	"food-api/models"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const bandulDateLayout = "2006-01-02"

var bandulPhonePattern = regexp.MustCompile(`^[0-9+\-\s()]*$`)

type BandulInput struct {
	DateJoin      string `json:"bandul_date_join"`
	DateBirth     string `json:"bandul_date_birth"`
	Name          string `json:"bandul_name"`
	Desc          string `json:"bandul_desc"`
	Phone         string `json:"bandul_phone"`
	MobilePhone   string `json:"bandul_mobile_phone"`
	ProvinceID    *uint  `json:"province_id"`
	CityID        *uint  `json:"city_id"`
	SubdistrictID *uint  `json:"subdistrict_id"`
	Address       string `json:"bandul_address"`
	Active        *bool  `json:"bandul_active"`
}

type BandulRow struct {
	ID              uint    `json:"bandul_id"`
	DateJoin        string  `json:"bandul_date_join"`
	DateBirth       *string `json:"bandul_date_birth"`
	Name            string  `json:"bandul_name"`
	Desc            string  `json:"bandul_desc"`
	Phone           string  `json:"bandul_phone"`
	MobilePhone     string  `json:"bandul_mobile_phone"`
	ProvinceID      *uint   `json:"province_id"`
	ProvinceName    string  `json:"province_name"`
	CityID          *uint   `json:"city_id"`
	CityName        string  `json:"city_name"`
	SubdistrictID   *uint   `json:"subdistrict_id"`
	SubdistrictName string  `json:"subdistrict_name"`
	Address         string  `json:"bandul_address"`
	Active          bool    `json:"bandul_active"`
}

func toBandulRow(b models.Bandul) BandulRow {
	row := BandulRow{
		ID:            b.ID,
		DateJoin:      b.DateJoin.UTC().Format(bandulDateLayout),
		Name:          b.Name,
		Desc:          b.Desc,
		Phone:         b.Phone,
		MobilePhone:   b.MobilePhone,
		ProvinceID:    b.IDProvince,
		CityID:        b.IDCity,
		SubdistrictID: b.IDSubdistrict,
		Address:       b.Address,
		Active:        b.Active == 1,
	}

	if b.DateBirth != nil {
		s := b.DateBirth.UTC().Format(bandulDateLayout)
		row.DateBirth = &s
	}
	if b.Province != nil {
		row.ProvinceName = b.Province.Name
	}
	if b.City != nil {
		row.CityName = strings.TrimSpace(b.City.Type + " " + b.City.Name)
	}
	if b.Subdistrict != nil {
		row.SubdistrictName = b.Subdistrict.Name
	}

	return row
}

func loadBandul(id uint) (models.Bandul, error) {
	var b models.Bandul
	err := database.DB.Preload("Province").Preload("City").Preload("Subdistrict").First(&b, id).Error
	return b, err
}

func findBandul(c *gin.Context, userID uint) (models.Bandul, bool) {
	var b models.Bandul

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err == nil {
		err = database.DB.Where("bandul_id = ? AND user_id = ?", id, userID).First(&b).Error
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Bandul tidak ditemukan",
		})
		return b, false
	}

	return b, true
}

func validasiBandul(in *BandulInput) (time.Time, *time.Time, map[string]string, error) {
	errs := map[string]string{}

	var join time.Time
	var birth *time.Time

	in.DateJoin = strings.TrimSpace(in.DateJoin)
	in.DateBirth = strings.TrimSpace(in.DateBirth)
	in.Name = strings.TrimSpace(in.Name)
	in.Desc = strings.TrimSpace(in.Desc)
	in.Phone = strings.TrimSpace(in.Phone)
	in.MobilePhone = strings.TrimSpace(in.MobilePhone)
	in.Address = strings.TrimSpace(in.Address)
	in.ProvinceID = zeroToNil(in.ProvinceID)
	in.CityID = zeroToNil(in.CityID)
	in.SubdistrictID = zeroToNil(in.SubdistrictID)

	if in.Name == "" {
		errs["bandul_name"] = "Nama bandul wajib diisi"
	} else if len([]rune(in.Name)) > 256 {
		errs["bandul_name"] = "Nama bandul maksimal 256 karakter"
	}

	if in.DateJoin == "" {
		errs["bandul_date_join"] = "Tanggal bergabung wajib diisi"
	} else if t, err := time.Parse(bandulDateLayout, in.DateJoin); err != nil {
		errs["bandul_date_join"] = "Tanggal bergabung tidak valid"
	} else {
		join = t
	}

	if in.DateBirth != "" {
		t, err := time.Parse(bandulDateLayout, in.DateBirth)
		if err != nil {
			errs["bandul_date_birth"] = "Tanggal lahir tidak valid"
		} else if t.After(time.Now()) {
			errs["bandul_date_birth"] = "Tanggal lahir tidak boleh di masa depan"
		} else {
			birth = &t
		}
	}

	if len(in.Phone) > 50 || !bandulPhonePattern.MatchString(in.Phone) {
		errs["bandul_phone"] = "Nomor telepon tidak valid"
	}
	if len(in.MobilePhone) > 50 || !bandulPhonePattern.MatchString(in.MobilePhone) {
		errs["bandul_mobile_phone"] = "Nomor HP tidak valid"
	}

	if in.CityID != nil && in.ProvinceID == nil {
		errs["province_id"] = "Provinsi wajib dipilih jika kota diisi"
	}
	if in.SubdistrictID != nil && in.CityID == nil {
		errs["city_id"] = "Kota wajib dipilih jika kecamatan diisi"
	}

	if in.ProvinceID != nil && errs["province_id"] == "" {
		found, err := wilayahAda("province", "province_id = ?", *in.ProvinceID)
		if err != nil {
			return join, birth, errs, err
		}
		if !found {
			errs["province_id"] = "Provinsi tidak ditemukan"
		}
	}

	if in.CityID != nil && in.ProvinceID != nil && errs["province_id"] == "" && errs["city_id"] == "" {
		found, err := wilayahAda("city", "city_id = ? AND province_id = ?", *in.CityID, *in.ProvinceID)
		if err != nil {
			return join, birth, errs, err
		}
		if !found {
			errs["city_id"] = "Kota tidak ditemukan di provinsi yang dipilih"
		}
	}

	if in.SubdistrictID != nil && in.CityID != nil && errs["city_id"] == "" && errs["subdistrict_id"] == "" {
		found, err := wilayahAda("subdistrict", "subdistrict_id = ? AND city_id = ?", *in.SubdistrictID, *in.CityID)
		if err != nil {
			return join, birth, errs, err
		}
		if !found {
			errs["subdistrict_id"] = "Kecamatan tidak ditemukan di kota yang dipilih"
		}
	}

	return join, birth, errs, nil
}

// GET /bandul?page=&limit=&q=&active=true|false
func GetBandul(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	q := database.DB.Model(&models.Bandul{}).Where("user_id = ?", userID)

	if s := strings.TrimSpace(c.Query("q")); s != "" {
		like := "%" + strings.ToLower(s) + "%"
		q = q.Where("(LOWER(bandul_name) LIKE ? OR bandul_phone LIKE ? OR bandul_mobile_phone LIKE ?)", like, like, like)
	}

	switch c.Query("active") {
	case "true":
		q = q.Where("bandul_active = 1")
	case "false":
		q = q.Where("bandul_active <> 1")
	}

	q = q.Session(&gorm.Session{})

	var total int64
	if err := q.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menghitung data bandul",
		})
		return
	}

	list := []models.Bandul{}
	err = q.Preload("Province").Preload("City").Preload("Subdistrict").
		Order("bandul_name ASC, bandul_id ASC").
		Limit(limit).
		Offset((page - 1) * limit).
		Find(&list).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal ambil data bandul",
		})
		return
	}

	rows := make([]BandulRow, 0, len(list))
	for _, b := range list {
		rows = append(rows, toBandulRow(b))
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        rows,
		"currentPage": page,
		"limit":       limit,
		"totalData":   total,
		"totalPages":  (int(total) + limit - 1) / limit,
	})
}

// GET /bandul/:id
func GetBandulBy(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}

	found, ok := findBandul(c, userID)
	if !ok {
		return
	}

	b, err := loadBandul(found.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal ambil data bandul",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": toBandulRow(b),
	})
}

// POST /bandul
func CreateBandul(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}

	var in BandulInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "JSON tidak valid",
		})
		return
	}

	join, birth, errs, err := validasiBandul(&in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal memeriksa data wilayah",
		})
		return
	}
	if len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": errs,
		})
		return
	}

	b := models.Bandul{
		UserID:        userID,
		DateJoin:      join,
		DateBirth:     birth,
		Name:          in.Name,
		Desc:          in.Desc,
		Phone:         in.Phone,
		MobilePhone:   in.MobilePhone,
		IDProvince:    in.ProvinceID,
		IDCity:        in.CityID,
		IDSubdistrict: in.SubdistrictID,
		Address:       in.Address,
		Active:        1,
	}
	if in.Active != nil {
		b.Active = activeToInt(*in.Active)
	}

	if err := database.DB.Create(&b).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menyimpan bandul",
		})
		return
	}

	saved, err := loadBandul(b.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Tersimpan, tapi gagal memuat ulang data",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": toBandulRow(saved),
	})
}

// PUT /bandul/:id
func UpdateBandul(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}

	b, found := findBandul(c, userID)
	if !found {
		return
	}

	var in BandulInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "JSON tidak valid",
		})
		return
	}

	join, birth, errs, err := validasiBandul(&in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal memeriksa data wilayah",
		})
		return
	}
	if len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": errs,
		})
		return
	}

	b.DateJoin = join
	b.DateBirth = birth
	b.Name = in.Name
	b.Desc = in.Desc
	b.Phone = in.Phone
	b.MobilePhone = in.MobilePhone
	b.IDProvince = in.ProvinceID
	b.IDCity = in.CityID
	b.IDSubdistrict = in.SubdistrictID
	b.Address = in.Address
	if in.Active != nil {
		b.Active = activeToInt(*in.Active)
	}

	if err := database.DB.Save(&b).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal mengubah bandul",
		})
		return
	}

	saved, err := loadBandul(b.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Tersimpan, tapi gagal memuat ulang data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": toBandulRow(saved),
	})
}

// DELETE /bandul/:id
func DeleteBandul(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}

	b, found := findBandul(c, userID)
	if !found {
		return
	}

	if err := database.DB.Delete(&b).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menghapus bandul",
		})
		return
	}

	c.Status(http.StatusNoContent)
}
