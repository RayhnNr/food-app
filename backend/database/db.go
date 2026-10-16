package database

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(){
	dsn := "host=localhost user=postgres password=Tpm12345678!@##@! dbname=food port=5432 sslmode=disable TimeZone=Asia/Jakarta"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal Konek Database: ", err)
	}

	DB = db
	log.Println("Database Terkoneksi")
}