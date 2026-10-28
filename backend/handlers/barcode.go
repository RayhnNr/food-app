package handlers

import (
	"errors"
	"fmt"
	"food-api/database"
	"food-api/middlewares"
	"food-api/models"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	barcodeMaxSerial   = 999999
	barcodeMaxPerBatch = 10000
	barcodeMaxCodeLen  = 32

	// defaultCompanyCode dipakai kalau env COMPANY_CODE kosong.
	// Nanti kalau sudah multi-company, ganti isi barcodeCompanyCode().
	defaultCompanyCode = "TPM"
)

var (
	errBarcodeTaken    = errors.New("kode barcode sudah ada")
	companyCodePattern = regexp.MustCompile(`^[A-Za-z0-9]{1,16}$`)
)

type BarcodeInternalInput struct {
	GoodsPartaiID *uint `json:"goods_partai_id"`
	Tahun         *int  `json:"tahun"`
	CodeFrom      *int  `json:"item_code_from"`
	CodeTo        *int  `json:"item_code_to"`
}

type BarcodeExternalInput struct {
	Header        string `json:"header"`
	Tahun         *int   `json:"tahun"`
	Digit         *int   `json:"item_code_digit"`
	GoodsPartaiID *uint  `json:"goods_partai_id"`
	CodeFrom      *int   `json:"item_code_from"`
	CodeTo        *int   `json:"item_code_to"`
}

type BarcodeBatchRow struct {
	ItemID          uint      `json:"item_id"`
	GoodsPartaiName string    `json:"goods_partai_name"`
	ItemType        string    `json:"item_type"`
	ItemYear        int       `json:"item_year"`
	ItemDateCreate  time.Time `json:"item_date_create"`
	CodeMin         string    `json:"code_min"`
	CodeMax         string    `json:"code_max"`
	Total           int64     `json:"total"`
}

type BarcodePrintRow struct {
	ItemCode        string `json:"item_code"`
	ItemType        string `json:"item_type"`
	GoodsPartaiName string `json:"goods_partai_name"`
	YearCode        string `json:"year_code"`
	CompanyCode     string `json:"company_code"`
	Serial          string `json:"serial"`
}

type barcodeResult struct {
	ItemID    uint   `json:"item_id"`
	Total     int    `json:"total"`
	CodeFirst string `json:"code_first"`
	CodeLast  string `json:"code_last"`
}

// tahunKeHuruf: dua digit terakhir tahun, tiap digit diganti huruf (0=J, 1=A, ... 9=I). 2026 menjadi BF.
func tahunKeHuruf(year int) string {
	const huruf = "JABCDEFGHI"
	yy := year % 100
	return string([]byte{huruf[yy/10], huruf[yy%10]})
}

func buildInternalCode(yearCode, company string, partaiCode, serial int) string {
	return fmt.Sprintf("%s-%s-%03d-%06d", yearCode, company, partaiCode, serial)
}

func buildExternalCode(header string, digit, serial int) string {
	return fmt.Sprintf("%s%0*d", header, digit, serial)
}

func barcodeDigitLimit(digit int) int {
	limit := 1
	for i := 0; i < digit; i++ {
		limit *= 10
	}
	return limit - 1
}

func potongKode(s string, start, length int) string {
	if start >= len(s) {
		return ""
	}

	end := start + length
	if end > len(s) {
		end = len(s)
	}

	return s[start:end]
}

// barcodePecahKode memecah kode untuk label: internal dengan tanda "-", external berdasarkan posisi.
func barcodePecahKode(itemType, code string) (year, company, serial string) {
	if itemType == models.ItemTypeInternal {
		parts := strings.Split(code, "-")
		if len(parts) != 4 {
			return "", "", ""
		}
		return parts[0], parts[1], parts[3]
	}

	return potongKode(code, 0, 2), potongKode(code, 2, 2), potongKode(code, 4, 5)
}

// barcodeCompanyCode mengambil kode company dari env COMPANY_CODE, kalau kosong pakai default TPM.
func barcodeCompanyCode() (string, bool) {
	code := strings.TrimSpace(os.Getenv("COMPANY_CODE"))
	if code == "" {
		code = defaultCompanyCode
	}
	return code, companyCodePattern.MatchString(code)
}

// nextInternalSerial mencari nomor urut internal berikutnya untuk satu partai pada satu tahun.
func nextInternalSerial(tx *gorm.DB, partaiID uint, year int) (int, error) {
	var last int

	err := tx.Raw(`SELECT COALESCE(MAX(CASE
			WHEN split_part(item_code, '-', 4) ~ '^[0-9]{1,9}$'
			THEN CAST(split_part(item_code, '-', 4) AS INTEGER)
		END), 0)
		FROM barcode_item
		WHERE goods_partai_id = ? AND item_type = ? AND item_year = ?`,
		partaiID, models.ItemTypeInternal, year).Scan(&last).Error

	return last + 1, err
}

func validasiBarcodeInternal(in *BarcodeInternalInput, userID uint) (models.GoodsPartai, map[string]string) {
	var partai models.GoodsPartai
	errs := map[string]string{}

	if in.GoodsPartaiID == nil || *in.GoodsPartaiID == 0 {
		errs["goods_partai_id"] = "Partai wajib diisi"
	}

	if in.Tahun == nil {
		errs["tahun"] = "Tahun wajib diisi"
	} else if *in.Tahun < 1990 || *in.Tahun > 2100 {
		errs["tahun"] = "Tahun tidak valid"
	}

	if in.CodeFrom == nil {
		errs["item_code_from"] = "Item kode dari wajib diisi"
	} else if *in.CodeFrom < 1 {
		errs["item_code_from"] = "Item kode dari minimal 1"
	}

	if in.CodeTo == nil {
		errs["item_code_to"] = "Item kode sampai wajib diisi"
	} else if in.CodeFrom != nil && *in.CodeTo < *in.CodeFrom {
		errs["item_code_to"] = "Item kode sampai tidak boleh lebih kecil dari item kode dari"
	} else if *in.CodeTo > barcodeMaxSerial {
		errs["item_code_to"] = fmt.Sprintf("Item kode sampai maksimal %d", barcodeMaxSerial)
	} else if in.CodeFrom != nil && *in.CodeTo-*in.CodeFrom+1 > barcodeMaxPerBatch {
		errs["item_code_to"] = fmt.Sprintf("Maksimal %d kode per generate", barcodeMaxPerBatch)
	}

	if len(errs) > 0 {
		return partai, errs
	}

	err := database.DB.Where("goods_partai_id = ? AND user_id = ?", *in.GoodsPartaiID, userID).First(&partai).Error
	if err != nil {
		errs["goods_partai_id"] = "Partai tidak ditemukan"
	} else if partai.Code < 1 || partai.Code > 999 {
		errs["goods_partai_id"] = "Kode partai harus 1 sampai 999 agar muat di kode barcode"
	}

	return partai, errs
}

func validasiBarcodeExternal(in *BarcodeExternalInput, userID uint) (models.GoodsPartai, map[string]string) {
	var partai models.GoodsPartai
	errs := map[string]string{}

	in.Header = strings.TrimSpace(in.Header)
	digitOK := in.Digit != nil && *in.Digit >= 1 && *in.Digit <= 9

	if in.Header == "" {
		errs["header"] = "Header wajib diisi"
	} else {
		for _, r := range in.Header {
			if r < 33 || r > 126 {
				errs["header"] = "Header hanya boleh berisi huruf, angka, dan simbol tanpa spasi"
				break
			}
		}
	}

	if in.Digit == nil {
		errs["item_code_digit"] = "Item kode digit wajib diisi"
	} else if !digitOK {
		errs["item_code_digit"] = "Item kode digit harus 1 sampai 9"
	} else if _, taken := errs["header"]; !taken && len(in.Header)+*in.Digit > barcodeMaxCodeLen {
		errs["header"] = fmt.Sprintf("Header terlalu panjang, header + digit maksimal %d karakter", barcodeMaxCodeLen)
	}

	if in.Tahun == nil {
		errs["tahun"] = "Tahun wajib diisi"
	} else if *in.Tahun < 1990 || *in.Tahun > 2100 {
		errs["tahun"] = "Tahun tidak valid"
	}

	if in.GoodsPartaiID == nil || *in.GoodsPartaiID == 0 {
		errs["goods_partai_id"] = "Partai wajib diisi"
	}

	if in.CodeFrom == nil {
		errs["item_code_from"] = "Item kode dari wajib diisi"
	} else if *in.CodeFrom < 1 {
		errs["item_code_from"] = "Item kode dari minimal 1"
	}

	if in.CodeTo == nil {
		errs["item_code_to"] = "Item kode sampai wajib diisi"
	} else if in.CodeFrom != nil && *in.CodeTo < *in.CodeFrom {
		errs["item_code_to"] = "Item kode sampai tidak boleh lebih kecil dari item kode dari"
	} else if digitOK && *in.CodeTo > barcodeDigitLimit(*in.Digit) {
		errs["item_code_to"] = fmt.Sprintf("Item kode sampai maksimal %d untuk %d digit", barcodeDigitLimit(*in.Digit), *in.Digit)
	} else if in.CodeFrom != nil && *in.CodeTo-*in.CodeFrom+1 > barcodeMaxPerBatch {
		errs["item_code_to"] = fmt.Sprintf("Maksimal %d kode per generate", barcodeMaxPerBatch)
	}

	if len(errs) > 0 {
		return partai, errs
	}

	err := database.DB.Where("goods_partai_id = ? AND user_id = ?", *in.GoodsPartaiID, userID).First(&partai).Error
	if err != nil {
		errs["goods_partai_id"] = "Partai tidak ditemukan"
	}

	return partai, errs
}

// insertBarcodeBatch menyimpan satu batch dalam satu transaksi. Partai dikunci supaya dua generate
// pada partai yang sama berjalan bergantian, dan rentang yang sudah terpakai ditolak.
func insertBarcodeBatch(userID, partaiID uint, itemType string, year, from, to int, build func(int) string) (barcodeResult, error) {
	var result barcodeResult

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		lock := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("goods_partai_id")
		if err := lock.First(&models.GoodsPartai{}, partaiID).Error; err != nil {
			return err
		}

		first, last := build(from), build(to)

		var taken int64
		err := tx.Model(&models.BarcodeItem{}).
			Where(`item_code COLLATE "C" BETWEEN ? AND ? AND LENGTH(item_code) = ?`, first, last, len(first)).
			Count(&taken).Error
		if err != nil {
			return err
		}
		if taken > 0 {
			return errBarcodeTaken
		}

		var batchID uint
		if err := tx.Raw("SELECT nextval('barcode_item_batch_seq')").Scan(&batchID).Error; err != nil {
			return err
		}

		items := make([]models.BarcodeItem, 0, to-from+1)
		for x := from; x <= to; x++ {
			items = append(items, models.BarcodeItem{
				ItemCode:       build(x),
				UserID:         userID,
				ItemYear:       year,
				ItemID:         batchID,
				ItemCodeStatus: 1,
				IDGoodsPartai:  partaiID,
				ItemType:       itemType,
			})
		}

		if err := tx.CreateInBatches(items, 1000).Error; err != nil {
			return err
		}

		result = barcodeResult{ItemID: batchID, Total: len(items), CodeFirst: first, CodeLast: last}
		return nil
	})

	return result, err
}

func respondBarcodeBatch(c *gin.Context, userID, partaiID uint, itemType string, year, from, to int, build func(int) string) {
	result, err := insertBarcodeBatch(userID, partaiID, itemType, year, from, to, build)
	if errors.Is(err, errBarcodeTaken) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": map[string]string{
				"item_code_from": "Sebagian kode pada rentang ini sudah ada",
			},
		})
		return
	}
	if err != nil {
		log.Printf("generate barcode: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menyimpan barcode",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": result,
	})
}

func barcodeItemID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("item_id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Barcode tidak ditemukan",
		})
		return 0, false
	}

	return uint(id), true
}

// GET /barcode
func GetBarcode(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}

	rows := []BarcodeBatchRow{}
	err := database.DB.Table("barcode_item AS i").
		Select(`i.item_id AS item_id,
			MAX(COALESCE(p.goods_partai_name, '')) AS goods_partai_name,
			MAX(i.item_type) AS item_type,
			MAX(i.item_year) AS item_year,
			MAX(i.item_date_create) AS item_date_create,
			MIN(i.item_code) AS code_min,
			MAX(i.item_code) AS code_max,
			COUNT(*) AS total`).
		Joins("LEFT JOIN goods_partai AS p ON p.goods_partai_id = i.goods_partai_id").
		Where("i.user_id = ?", userID).
		Group("i.item_id").
		Order("i.item_id DESC").
		Scan(&rows).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal ambil data barcode",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": rows,
	})
}

// GET /barcode/:item_id
func GetBarcodeDetail(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}

	itemID, ok := barcodeItemID(c)
	if !ok {
		return
	}

	items := []models.BarcodeItem{}
	err := database.DB.Where("item_id = ? AND user_id = ?", itemID, userID).
		Order("item_code").
		Find(&items).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal ambil data barcode",
		})
		return
	}
	if len(items) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Barcode tidak ditemukan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": items,
	})
}

// GET /barcode/next-serial?goods_partai_id=..&tahun=..
func GetBarcodeNextSerial(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}

	errs := map[string]string{}

	partaiID, err := strconv.ParseUint(c.Query("goods_partai_id"), 10, 64)
	if err != nil || partaiID == 0 {
		errs["goods_partai_id"] = "Partai wajib diisi"
	}

	tahun, err := strconv.Atoi(c.Query("tahun"))
	if err != nil || tahun < 1990 || tahun > 2100 {
		errs["tahun"] = "Tahun tidak valid"
	}

	if len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": errs,
		})
		return
	}

	err = database.DB.Select("goods_partai_id").
		Where("goods_partai_id = ? AND user_id = ?", partaiID, userID).
		First(&models.GoodsPartai{}).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Partai tidak ditemukan",
		})
		return
	}

	next, err := nextInternalSerial(database.DB, uint(partaiID), tahun)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menghitung nomor urut",
		})
		return
	}
	if next > barcodeMaxSerial {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Kode sudah penuh",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"next": next,
	})
}

// POST /barcode/internal
func GenerateBarcodeInternal(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}

	var in BarcodeInternalInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "JSON tidak valid",
		})
		return
	}

	company, ok := barcodeCompanyCode()
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Kode company belum diatur di server",
		})
		return
	}

	partai, errs := validasiBarcodeInternal(&in, userID)
	if len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": errs,
		})
		return
	}

	yearCode := tahunKeHuruf(*in.Tahun)
	build := func(serial int) string {
		return buildInternalCode(yearCode, company, partai.Code, serial)
	}

	respondBarcodeBatch(c, userID, partai.ID, models.ItemTypeInternal, *in.Tahun, *in.CodeFrom, *in.CodeTo, build)
}

// POST /barcode/external
func GenerateBarcodeExternal(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}

	var in BarcodeExternalInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "JSON tidak valid",
		})
		return
	}

	partai, errs := validasiBarcodeExternal(&in, userID)
	if len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": errs,
		})
		return
	}

	digit := *in.Digit
	build := func(serial int) string {
		return buildExternalCode(in.Header, digit, serial)
	}

	respondBarcodeBatch(c, userID, partai.ID, models.ItemTypeExternal, *in.Tahun, *in.CodeFrom, *in.CodeTo, build)
}

// DELETE /barcode/:item_id
func DeleteBarcode(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}

	itemID, ok := barcodeItemID(c)
	if !ok {
		return
	}

	var used int64
	err := database.DB.Model(&models.BarcodeItem{}).
		Where("item_id = ? AND user_id = ? AND item_code_status <> 1", itemID, userID).
		Count(&used).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menghapus barcode",
		})
		return
	}
	if used > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"message": "Sebagian barcode sudah dipakai, tidak bisa dihapus",
		})
		return
	}

	result := database.DB.Where("item_id = ? AND user_id = ?", itemID, userID).Delete(&models.BarcodeItem{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menghapus barcode",
		})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Barcode tidak ditemukan",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// GET /barcode/:item_id/print?item_code=..
func GetBarcodePrint(c *gin.Context) {
	userID, ok := middlewares.GetUserIDFromCOntext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User tidak dikenali",
		})
		return
	}

	itemID, ok := barcodeItemID(c)
	if !ok {
		return
	}

	q := database.DB.Table("barcode_item AS i").
		Select(`i.item_code AS item_code,
			i.item_type AS item_type,
			COALESCE(p.goods_partai_name, '') AS goods_partai_name`).
		Joins("LEFT JOIN goods_partai AS p ON p.goods_partai_id = i.goods_partai_id").
		Where("i.item_id = ? AND i.user_id = ?", itemID, userID)

	if code := strings.TrimSpace(c.Query("item_code")); code != "" {
		q = q.Where("i.item_code = ?", code)
	}

	rows := []BarcodePrintRow{}
	if err := q.Order("i.item_code").Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal ambil data cetak",
		})
		return
	}
	if len(rows) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Barcode tidak ditemukan",
		})
		return
	}

	for i := range rows {
		rows[i].YearCode, rows[i].CompanyCode, rows[i].Serial = barcodePecahKode(rows[i].ItemType, rows[i].ItemCode)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": rows,
	})
}