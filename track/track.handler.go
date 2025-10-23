package track

import (
	"fmt"
	"log"
	"path/filepath" // Import baru untuk mendapatkan ekstensi file
	"valeth-soundcloud-api/config"

	"github.com/gofiber/fiber/v2"
	storage_go "github.com/supabase-community/storage-go"
	"gorm.io/gorm"
)

// Handler struct memegang semua "koneksi" yang dibutuhkan
// (DB, Supabase, dan Config)
type Handler struct {
	DB                 *gorm.DB
	Supabase           *storage_go.Client
	Config             config.Config
}

// NewHandler adalah "factory" untuk membuat handler baru
// Ini akan kita panggil dari main.go nanti
func NewHandler(db *gorm.DB, supabase *storage_go.Client, config config.Config) *Handler {
	return &Handler{
		DB:                 db,
		Supabase:           supabase,
		Config:             config,
	}
}

// UploadTrack adalah fungsi yang menangani logika upload
func (h *Handler) UploadTrack(c *fiber.Ctx) error {
	// 1. Parse form-data
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Gagal parse form",
		})
	}

	// 2. Ambil data teks (title & artist)
	title := form.Value["title"]
	artist := form.Value["artist"]
	if len(title) == 0 || len(artist) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Title dan Artist wajib diisi",
		})
	}

	// 3. Ambil file MP3 (track_file)
	files := form.File["track_file"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File track (track_file) wajib diisi",
		})
	}
	trackFileHeader := files[0]

	// 4. Ambil file Gambar (cover_file)
	coverFiles := form.File["cover_file"]
	if len(coverFiles) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File cover (cover_file) wajib diisi",
		})
	}
	coverFileHeader := coverFiles[0]

	// --- PROSES UPLOAD TRACK (MP3) ---
	trackFile, err := trackFileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuka file track"})
	}
	defer trackFile.Close()

	// Buat nama file unik (misal: 1678886400-track.mp3)
	// Kita akan menggunakan Request ID yang nanti kita set di main.go
	trackFileName := fmt.Sprintf("%v-%s", c.Locals("requestid"), trackFileHeader.Filename)
	trackContentType := trackFileHeader.Header.Get("Content-Type")

	// Upload ke Supabase Storage (Bucket 'tracks')
	// <-- PERBAIKAN 1: Menambahkan '&'
	_, err = h.Supabase.UploadFile(h.Config.SUPABASE_BUCKET_tracks, trackFileName, trackFile, storage_go.FileOptions{
		ContentType: &trackContentType,
	})
	if err != nil {
		log.Println("Error upload track:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal upload file track"})
	}
	log.Println("Berhasil upload file track:", trackFileName)

	// --- PROSES UPLOAD COVER (GAMBAR) ---
	coverFile, err := coverFileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuka file cover"})
	}
	defer coverFile.Close()

	// Buat nama file unik (misal: 1678886400-cover.jpg)
	coverExt := filepath.Ext(coverFileHeader.Filename) // Ambil ekstensi file (misal: .jpg)
	coverFileName := fmt.Sprintf("%v-cover%s", c.Locals("requestid"), coverExt)
	coverContentType := coverFileHeader.Header.Get("Content-Type")

	// Upload ke Supabase Storage (Bucket 'covers')
	// <-- PERBAIKAN 2: Menambahkan '&'
	_, err = h.Supabase.UploadFile(h.Config.SUPABASE_BUCKET_covers, coverFileName, coverFile, storage_go.FileOptions{
		ContentType: &coverContentType,
	})
	if err != nil {
		log.Println("Error upload cover:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal upload file cover"})
	}
	log.Println("Berhasil upload file cover:", coverFileName)


	// 5. Dapatkan URL publik untuk kedua file
	trackURL := h.Supabase.GetPublicUrl(h.Config.SUPABASE_BUCKET_tracks, trackFileName)
	coverURL := h.Supabase.GetPublicUrl(h.Config.SUPABASE_BUCKET_covers, coverFileName)

	// 6. Simpan data ke Database PostgreSQL
	newTrack := Track{
		Title:    title[0],
		Artist:   artist[0],
		URL:      trackURL.SignedURL, // URL untuk MP3
		PublicID: trackFileName,      // Nama file MP3 (untuk hapus)
		CoverURL: coverURL.SignedURL, // URL untuk Gambar
	}

	if result := h.DB.Create(&newTrack); result.Error != nil {
		// Jika gagal simpan DB, hapus file yg sudah diupload (rollback)
		// <-- PERBAIKAN 3: Mengganti DeleteFile menjadi RemoveFile
		h.Supabase.RemoveFile(h.Config.SUPABASE_BUCKET_tracks, []string{trackFileName})
		// <-- PERBAIKAN 4: Mengganti DeleteFile menjadi RemoveFile
		h.Supabase.RemoveFile(h.Config.SUPABASE_BUCKET_covers, []string{coverFileName})
		
		log.Println("Error simpan ke DB:", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan data track"})
	}

	// 7. Kembalikan data track yang baru sebagai JSON
	return c.Status(fiber.StatusCreated).JSON(newTrack)
}

