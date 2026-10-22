package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const maxUpload = 2 << 20

var tipeBoleh = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

func UploadGambar(c *gin.Context){
	file, err := c.FormFile("gambar")

	// Cek wajib diisi
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors" : map[string]string{
				"gambar" : "File gambar wajib dipilih!",
			},
		})
		return
	}

	// Validasi size
	if file.Size > maxUpload {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors" : map[string]string{
				"gambar" : "Size file maksimal 2MB",
			},
		})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message" : "Gagal membaca file",
		})
		return
	}

	defer src.Close()

	buf := make([]byte, 512)
	n,_ := src.Read(buf)
	tipe := http.DetectContentType(buf[:n])

	ext, oke := tipeBoleh[tipe]

	// Validasi tipe file
	if !oke {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors" : map[string]string{
				"gambar" : "Hanya JPG, PNG, atau WEBP",
			},
		})
		return
	}

	nama := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	tujuan := filepath.Join("uploads", nama)

	if err := os.MkdirAll("uploads", 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message" : "Gagal menyiapkan file",
		})
		return
	}

	if err := c.SaveUploadedFile(file, tujuan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message" : "Gagal menyimpan file",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"gambar": nama,
		"url":    "/uploads/" + strings.TrimPrefix(nama, "/"),
	})





}