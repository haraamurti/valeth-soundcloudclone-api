package database

import (
	"log"
	"valeth-soundcloud-api/config" // Mengimpor paket config yang Anda buat

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectDatabase menggunakan Config struct untuk terhubung ke DB.
func ConnectDatabase(config config.Config) (*gorm.DB, error) {
	// Di sini kita menggunakan nilai yang sudah dibaca dari .env
	// yaitu "db_URL" yang Anda petakan ke DatabaseURL
	dsn := config.DatabaseURL

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal terhubung ke Database")
		return nil, err
	}

	return db, nil
}

