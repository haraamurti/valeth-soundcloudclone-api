package main

import (
	"log"
	"valeth-soundcloud-api/config"
	"valeth-soundcloud-api/database"
	"valeth-soundcloud-api/storage" // Paket storage Anda
	"valeth-soundcloud-api/track"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	// --- PERBAIKAN 1: Import library RESMI yang baru (dari nedpals) ---
	supabase "github.com/nedpals/supabase-go"
	"gorm.io/gorm"
	// --- Library 'storage_go' yang lama sudah dihapus ---
)

// --- PERBAIKAN 2: Gunakan tipe client yang baru ---
type AppState struct {
	DB       *gorm.DB
	Supabase *supabase.Client // <-- TIPE BARU
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

	// --- PERBAIKAN 3 (KRITIS): Tangani 2 nilai kembalian (client, error) ---
	supabaseClient, err := storage.InitSupabase(config)
	if err != nil {
		// Jika URL/Key salah, kita akan berhenti di sini
		log.Fatal("Gagal menginisialisasi client Supabase: ", err)
	}
	// --- AKHIR PERBAIKAN 3 ---

	// --- ⬇⬇⬇ INI PERBAIKANNYA ⬇⬇⬇ ---
	// inisialisasi fiber dengan BodyLimit yang lebih besar
	app := fiber.New(fiber.Config{
		// Set batas upload ke 100 MB (100 * 1024 * 1024)
		BodyLimit: 100 * 1024 * 1024,
	})
	// --- ⬆⬆⬆ AKHIR PERBAIKAN ⬆⬆⬆ ---

	app.Use(logger.New()) // menambah logger untuk tiap request
	app.Use(requestid.New())
	//setuping rute
	v1 := app.Group("/api/v1")

	//mencoba ping route link ke localhost dengan/
	v1.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Selamat datang di valeth SoundCloud API !",
		})
	})

	// <-- BARIS BARU 1: Buat instance Handler-nya
	// Sekarang `supabaseClient` memiliki tipe yang benar
	trackHandler := track.NewHandler(DB, supabaseClient, config)

	// <-- BARIS BARU 2: Daftarkan semua rute track (dari track.routes.go)
	track.SetupTrackRoutes(v1, trackHandler)
	
	// jalankan server
	log.Println("Server berjalan di port 1975...")
	log.Fatal(app.Listen(":1975")) // (Saya biarkan port 1975 Anda)
}

