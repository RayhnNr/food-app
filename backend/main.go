package main

import (
	"food-api/database"
	"food-api/models"
	"food-api/handlers"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	database.Connect()
	database.DB.AutoMigrate(
		&models.Product{},
		&models.Keranjang{},
		&models.Pesanan{},
		&models.PesananItem{},
	)
	r := gin.Default()

	// CORS: izinkan Vue (Vite) memanggil API ini
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type"},
	}))

	// Route
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// r.POST("/products", handlers.CreateProduct)
	r.GET("/products", handlers.GetData)
	r.GET("/products/:id", handlers.GetProduct)
	r.POST("/products", handlers.CreateProduct)
	r.PUT("/products/:id", handlers.UpdateProduct)
	r.DELETE("/products/:id", handlers.DeleteProduct)

	// Keranjang
	r.GET("/keranjangs", handlers.GetKeranjang)
	r.POST("/keranjangs", handlers.TambahKeranjang)
	r.PATCH("/keranjangs/:id", handlers.UpdateKeranjang)
	r.DELETE("/keranjangs/:id", handlers.DeleteKeranjang)
	r.DELETE("/keranjangs", handlers.ClearKeranjang)

	// Pesanan
	// Endpoint admin
	admin := r.Group("/admin")
	{
		admin.GET("/pesanans", handlers.GetPesanans)
		admin.GET("/pesanans/:id", handlers.GetPesanan)
		admin.PUT("/pesanans/:id", handlers.UpdatePesanan)
		admin.DELETE("/pesanans/:id", handlers.HapusPesanan)
	}

	r.Run(":8080")
}