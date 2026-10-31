package main

import (
	"food-api/database"
	"food-api/handlers"
	"food-api/middlewares"
	"food-api/models"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env tidak ditemukan")
	}

	database.Connect()
	database.DB.AutoMigrate(
		&models.Category{},
		&models.Product{},
		&models.Keranjang{},
		&models.Pesanan{},
		&models.PesananItem{},
		&models.User{}, 
		&models.CoaType{}, 
		&models.Coa{},
		&models.CoaJournalType{},
		&models.CoaJournal{},
		&models.CoaGoods{},
		&models.GoodsPartai{},
		&models.GoodsCoaPartai{},
		&models.GoodsPartaiDecimal{},
		&models.GoodsPartaiTarra{},
		&models.BarcodeItem{},
		&models.GoodsGrade{},
		&models.GoodsCategory{},
		&models.Goods{},
		&models.GoodsCoa{},
	)

	r := gin.Default()
	

	r.MaxMultipartMemory = 5 << 20
	r.Static("/uploads", "./uploads")

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"}, // <- tambah Authorization
	}))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// ===== PUBLIC =====
	r.POST("/auth/register", handlers.Register)
	r.POST("/auth/login", handlers.Login)
	r.GET("/products", handlers.GetData)
	r.GET("/products/:id", handlers.GetProduct)

	// ===== PROTECTED (wajib login) =====
	auth := r.Group("/")
	auth.Use(middlewares.AuthMiddleware())
	{
		auth.GET("/profile", handlers.GetProfile)

		// Produk (tulis)
		auth.POST("/products", handlers.CreateProduct)
		auth.PUT("/products/:id", handlers.UpdateProduct)
		auth.DELETE("/products/:id", handlers.DeleteProduct)
		auth.POST("/upload", handlers.UploadGambar)

		// Category
		auth.GET("/category", handlers.GetCategory)
		auth.POST("/category", handlers.CreateCategory)
		auth.PUT("/category/:id", handlers.UpdateCategory)
		auth.DELETE("/category/:id", handlers.DeleteCategory)

		// Keranjang
		auth.GET("/keranjangs", handlers.GetKeranjang)
		auth.POST("/keranjangs", handlers.TambahKeranjang)
		auth.PATCH("/keranjangs/:id", handlers.UpdateKeranjang)
		auth.DELETE("/keranjangs/:id", handlers.DeleteKeranjang)
		auth.DELETE("/keranjangs", handlers.ClearKeranjang)

		// Pesanan
		auth.GET("/pesanans", handlers.GetPesanans)
		auth.POST("/pesanans", handlers.TambahPesanan)

		// COA
		auth.GET("/coa-types", handlers.GetCoaTypes)
		auth.GET("/coa", handlers.GetCoa)
		auth.GET("/coa/next-code", handlers.GetNextCoaCode)
		auth.GET("/coa/tree", handlers.GetCoaTree)
		auth.GET("/coa/:id", handlers.GetCoaBy)
		auth.POST("/coa", handlers.CreateCoa)
		auth.PUT("/coa/:id", handlers.UpdateCoa)
		auth.DELETE("/coa/:id", handlers.DeleteCoa)

		// CoA Journal
		auth.GET("/coa-journal", handlers.GetCoaJurnal)
		auth.PUT("/coa-journal", handlers.SaveCoaJournal)

		// Goods Partai
		auth.GET("/coa-goods", handlers.GetCoaGoods)
		auth.GET("/goods-partai", handlers.GetGoodsPartai)
		auth.GET("/goods-partai/:id", handlers.GetGoodsPartaiBy)
		auth.POST("/goods-partai", handlers.CreateGoodsPartai)
		auth.PUT("/goods-partai/:id", handlers.UpdateGoodsPartai)
		auth.DELETE("/goods-partai/:id", handlers.DeleteGoodsPartai)

		auth.GET("/goods-partai/:id/tarra", handlers.GetGoodsPartaiTarra)
		auth.POST("/goods-partai/:id/tarra", handlers.CreateGoodsPartaiTarra)
		auth.DELETE("/goods-partai/:id/tarra/:tarra_id", handlers.DeleteGoodsPartaiTarra)

		auth.GET("/goods-partai/:id/decimal", handlers.GetGoodsPartaiDecimal)
		auth.POST("/goods-partai/:id/decimal", handlers.CreateGoodsPartaiDecimal)
		auth.DELETE("/goods-partai/:id/decimal/:decimal_id", handlers.DeleteGoodsPartaiDecimal)


		// Goods Grade
		auth.GET("/goods-grade", handlers.GetGoodsGrade)
		auth.GET("/goods-grade/:id", handlers.GetGoodsGradeBy)
		auth.POST("/goods-grade", handlers.CreateGoodsGrade)
		auth.PUT("/goods-grade/:id", handlers.UpdateGoodsGrade)
		auth.DELETE("/goods-grade/:id", handlers.DeleteGoodsGrade)


		// Goods
		auth.GET("/goods", handlers.GetGoods)
		auth.GET("/goods/:id", handlers.GetGoodsBy)
		auth.POST("/goods", handlers.CreateGoods)
		auth.PUT("/goods/:id", handlers.UpdateGoods)
		auth.DELETE("/goods/:id", handlers.DeleteGoods)
		auth.GET("/goods-category-options", handlers.GetGoodsCategoryOptions)
		auth.POST("/goods-categories", handlers.CreateGoodsCategory)

		// Barcode
		auth.GET("/barcode", handlers.GetBarcode)
		auth.GET("/barcode/next-serial", handlers.GetBarcodeNextSerial)
		auth.POST("/barcode/internal", handlers.GenerateBarcodeInternal)
		auth.POST("/barcode/external", handlers.GenerateBarcodeExternal)
		auth.GET("/barcode/:item_id", handlers.GetBarcodeDetail)
		auth.GET("/barcode/:item_id/print", handlers.GetBarcodePrint)
		auth.DELETE("/barcode/:item_id", handlers.DeleteBarcode)

		// Admin
		admin := auth.Group("/admin")
		{
			admin.GET("/pesanans", handlers.GetPesanans)
			admin.GET("/pesanans/:id", handlers.GetPesanan)
			admin.PUT("/pesanans/:id", handlers.UpdatePesanan)
			admin.DELETE("/pesanans/:id", handlers.HapusPesanan)
			
		}
	}

	r.Run(":8080")
}