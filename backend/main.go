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