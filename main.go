package main

import (
	"log"
	"valeth-soundcloud-api/config"
	"valeth-soundcloud-api/database"
	"valeth-soundcloud-api/storage"

	// Paket storage Anda
	"valeth-soundcloud-api/track"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	storage_go "github.com/supabase-community/storage-go"
	"gorm.io/gorm"
)

// Kita akan menyimpan objek DB dan Supabase di sini agar bisa diakses
// oleh handler nanti.
type AppState struct {
	DB       *gorm.DB
	Supabase *storage_go.Client
}

func main() {
	//bakal loading config kesini
	config, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("Gagal memuat konfigurasi .env: ", err)
	}

	//connect database dari data yang didapat dari config
	DB, err := database.ConnectDatabase(config)
	if err != nil {
		log.Fatal("Gagal terhubung ke database: ", err)
	}
	log.Println("Koneksi Database berhasil.")

	//automigration
	log.Println("Menjalankan Auto-Migration...")
	err = DB.AutoMigrate(&track.Track{})
	if err != nil {
		log.Fatal("Gagal melakukan Auto-Migration: ", err)
	}
	log.Println("Database berhasil di-Migrasi.")
	
	// inisialisasi fiber
	app := fiber.New()
	app.Use(logger.New()) // menambah logger untuk tiap request
	app.Use(requestid.New())
	//setuping rute
	v1 := app.Group("/api/v1")


	//mencoba ping route link ke localhost dengan/
	v1.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Selamat datang di valeth SoundCloud API !",
		})})

supabaseClient := storage.InitSupabase(config)
	// <-- BARIS BARU 1: Buat instance Handler-nya
	trackHandler := track.NewHandler(DB, supabaseClient, config)

	// <-- BARIS BARU 2: Daftarkan semua rute track (dari track.routes.go)
	track.SetupTrackRoutes(v1, trackHandler)
	// jalankan server
	log.Println("Server berjalan di port 7272...")
	log.Fatal(app.Listen(":7272"))
}

