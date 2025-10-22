package main

import (
	"log"
	"valeth-soundcloud-api/config"   // Paket config Anda
	"valeth-soundcloud-api/database" // Paket database Anda

	// Paket storage Anda
	"valeth-soundcloud-api/track" // Paket track Anda (untuk model)

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	storage_go "github.com/supabase-community/storage-go" // <-- IMPORT DITAMBAHKAN
	"gorm.io/gorm"
)

// Kita akan menyimpan objek DB dan Supabase di sini agar bisa diakses
// oleh handler nanti.
type AppState struct {
	DB       *gorm.DB
	Supabase *storage_go.Client
}

func main() {
	// 1. Memuat Konfigurasi
	// Kita memuat dari "." (folder saat ini)
	config, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("Gagal memuat konfigurasi .env: ", err)
	}

	// 2. Menghubungkan ke Database
	DB, err := database.ConnectDatabase(config)
	if err != nil {
		log.Fatal("Gagal terhubung ke database: ", err)
	}
	log.Println("Koneksi Database berhasil.")

	// 3. Menjalankan Auto-Migration
	// Ini adalah langkah KRUSIAL.
	// GORM akan melihat struct `track.Track` dan membuat
	// tabel `tracks` di database Supabase Anda secara otomatis.
	log.Println("Menjalankan Auto-Migration...")
	err = DB.AutoMigrate(&track.Track{})
	if err != nil {
		log.Fatal("Gagal melakukan Auto-Migration: ", err)
	}
	log.Println("Database berhasil di-Migrasi.")

	// 4. Menginisialisasi Klien Supabase
	//supabaseClient := storage.InitSupabase(config)

	// (Kita akan tambahkan ini nanti, tapi siapkan dulu)
	// appState := &AppState{
	// 	DB:       DB,
	// 	Supabase: supabaseClient,
	// }

	// 5. Menginisialisasi Fiber (Web Server)
	app := fiber.New()
	app.Use(logger.New()) // Menambahkan logger untuk setiap request

	// Rute percobaan
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Selamat datang di valeth SoundCloud API !", // <-- PESAN DIPERBARUI
		})
	}) // <-- ERROR SINTAKS DIPERBAIKI (ditambahkan ')' )

	// (Nanti kita akan tambahkan rute API di sini)
	// track.RegisterRoutes(api, appState)

	// 6. Menjalankan Server
	log.Println("Server berjalan di port 7272...")
	log.Fatal(app.Listen(":7272"))
}

