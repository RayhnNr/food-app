package handlers

import (
	"net/http"
	"strings"

	"food-api/database"
	"food-api/models"

	"github.com/gin-gonic/gin"
)

type PesananInput struct {
    Nama   string `json:"nama" binding:"required"`
    NoMeja int    `json:"no_meja" binding:"required"`
}

type UpdatePesananInput struct {
	Nama   string                   `json:"nama" binding:"required"`
	NoMeja int                      `json:"no_meja" binding:"required"`
	Status string                   `json:"status" binding:"required"`
	Items  []UpdatePesananItemInput `json:"items" binding:"required"`
}

type UpdatePesananItemInput struct {
	ID         uint   `json:"id" binding:"required"`
	Jumlah     int    `json:"jumlah" binding:"required"`
	Keterangan string `json:"keterangan"`
}

func GetPesanans(c *gin.Context) {
	pesanans := []models.Pesanan{}

	db := database.DB.Preload("Items").Order("id DESC")

	if status := strings.TrimSpace(c.Query("status")); status != "" {
		db = db.Where("status = ?", status)
	}
	if q := strings.TrimSpace(c.Query("q")); q != "" {
		db = db.Where("nama ILIKE ?", "%"+q+"%")
	}

	if err := db.Find(&pesanans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal ambil data pesanan"})
		return
	}

	c.JSON(http.StatusOK, pesanans)
}

func TambahPesanan(c *gin.Context) {
	var in PesananInput

	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "JSON tidak valid",
		})
		return
	}

	in.Nama = strings.TrimSpace(in.Nama)

	errs := map[string]string{}

	if in.Nama == "" {
		errs["nama"] = "Nama wajib diisi"
	}

	if in.NoMeja <= 0 {
		errs["no_meja"] = "Nomor meja wajib diisi"
	}

	if len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": errs,
		})
		return
	}

	// Mulai satu transaksi
	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal memulai transaksi",
		})
		return
	}

	// Ambil isi keranjang
	var keranjangs []models.Keranjang

	if err := tx.Preload("Product").Find(&keranjangs).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal mengambil barang",
		})
		return
	}

	if len(keranjangs) == 0 {
		tx.Rollback()
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Keranjang masih kosong",
		})
		return
	}

	pesanan := models.Pesanan{
		Nama:   in.Nama,
		NoMeja: in.NoMeja,
		Status: "diproses",
	}

	// Hitung total dan simpan detail barang
	for _, item := range keranjangs {
		pesanan.Total += item.Product.Harga * item.Jumlah

		pesanan.Items = append(pesanan.Items, models.PesananItem{
			ProductID:  item.ProductID,
			Nama:       item.Product.Nama,
			Harga:      item.Product.Harga,
			Jumlah:     item.Jumlah,
			Keterangan: item.Keterangan,
		})
	}

	// Simpan pesanan beserta item-nya
	if err := tx.Create(&pesanan).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menyimpan pesanan",
		})
		return
	}

	// Kosongkan keranjang
	if err := tx.Where("1 = 1").Delete(&models.Keranjang{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal mengosongkan keranjang",
		})
		return
	}

	// Selesaikan transaksi
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menyelesaikan pesanan",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Pesanan berhasil dibuat",
		"data":    pesanan,
	})
}

func GetPesanan(c *gin.Context) {
	var pesanan models.Pesanan

	if err := database.DB.Preload("Items").First(&pesanan, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Pesanan tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, pesanan)
}

func UpdatePesanan(c *gin.Context) {
	var in UpdatePesananInput

	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Data JSON tidak valid",
		})
		return
	}

	in.Nama = strings.TrimSpace(in.Nama)
	in.Status = strings.TrimSpace(in.Status)

	if in.Nama == "" || in.NoMeja <= 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Nama dan nomor meja wajib diisi",
		})
		return
	}

	if in.Status != "diproses" && in.Status != "selesai" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Status pesanan tidak valid",
		})
		return
	}

	if len(in.Items) == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Pesanan harus memiliki minimal satu item",
		})
		return
	}

	// Validasi jumlah dan ID item
	seen := make(map[uint]bool)

	for _, item := range in.Items {
		if item.ID == 0 || item.Jumlah <= 0 || seen[item.ID] {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"message": "ID item atau jumlah tidak valid",
			})
			return
		}
		seen[item.ID] = true
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal memulai transaksi",
		})
		return
	}

	var pesanan models.Pesanan

	if err := tx.First(&pesanan, c.Param("id")).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Pesanan tidak ditemukan",
		})
		return
	}

	// Ambil item yang memang dimiliki pesanan ini
	var itemsLama []models.PesananItem

	if err := tx.Where("pesanan_id = ?", pesanan.ID).
		Find(&itemsLama).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal mengambil item pesanan",
		})
		return
	}

	itemMap := make(map[uint]models.PesananItem)

	for _, item := range itemsLama {
		itemMap[item.ID] = item
	}

	total := 0
	ids := make([]uint, 0, len(in.Items))

	for _, input := range in.Items {
		item, ada := itemMap[input.ID]

		if !ada {
			tx.Rollback()
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"message": "Item tidak ditemukan dalam pesanan ini",
			})
			return
		}

		if err := tx.Model(&models.PesananItem{}).
			Where("id = ? AND pesanan_id = ?", item.ID, pesanan.ID).
			Updates(map[string]interface{}{
				"jumlah":     input.Jumlah,
				"keterangan": input.Keterangan,
			}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Gagal memperbarui item pesanan",
			})
			return
		}

		// Harga diambil dari database, bukan dari frontend
		total += item.Harga * input.Jumlah
		ids = append(ids, item.ID)
	}

	// Hapus item yang dihapus dari form frontend
	if err := tx.Where(
		"pesanan_id = ? AND id NOT IN ?",
		pesanan.ID,
		ids,
	).Delete(&models.PesananItem{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menghapus item pesanan",
		})
		return
	}

	// Perbarui data utama pesanan
	if err := tx.Model(&pesanan).Updates(map[string]interface{}{
		"nama":    in.Nama,
		"no_meja": in.NoMeja,
		"status":  in.Status,
		"total":   total,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal memperbarui pesanan",
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menyelesaikan transaksi",
		})
		return
	}

	// Ambil kembali data terbaru beserta item-nya
	if err := database.DB.Preload("Items").
		First(&pesanan, pesanan.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Pesanan tersimpan, tetapi gagal mengambil data terbaru",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pesanan berhasil diperbarui",
		"data":    pesanan,
	})
}

func HapusPesanan(c *gin.Context) {
	id := c.Param("id")

	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal memulai transaksi",
		})
		return
	}

	var pesanan models.Pesanan

	if err := tx.First(&pesanan, id).Error; err != nil {
		tx.Rollback()

		c.JSON(http.StatusNotFound, gin.H{
			"message": "Pesanan tidak ditemukan",
		})
		return
	}

	// Hapus detail item pesanan terlebih dahulu
	if err := tx.Where("pesanan_id = ?", pesanan.ID).
		Delete(&models.PesananItem{}).Error; err != nil {
		tx.Rollback()

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menghapus item pesanan",
		})
		return
	}

	// Hapus pesanan utama
	if err := tx.Delete(&pesanan).Error; err != nil {
		tx.Rollback()

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menghapus pesanan",
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menyelesaikan transaksi",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pesanan berhasil dihapus",
	})
}