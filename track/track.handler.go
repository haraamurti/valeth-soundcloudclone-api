package track

import (
	"fmt"
	"log"
	"path/filepath"
	"valeth-soundcloud-api/config"

	// <-- Kita masih butuh ini untuk handler lama

	"github.com/gofiber/fiber/v2"
	supabase "github.com/nedpals/supabase-go"
	"gorm.io/gorm"
)

// ... (struct Handler dan NewHandler Anda tetap sama) ...
type Handler struct {
	DB       *gorm.DB
	Supabase *supabase.Client
	Config   config.Config
}

func NewHandler(db *gorm.DB, supabase *supabase.Client, config config.Config) *Handler {
	return &Handler{
		DB:       db,
		Supabase: supabase,
		Config:   config,
	}
}

// ... (Fungsi UploadTrack Anda yang sudah benar tetap di sini) ...
func (h *Handler) UploadTrack(c *fiber.Ctx) error {
	// ...
	// (Seluruh kode UploadTrack Anda yang sudah ada tetap di sini)
	// ...
	// (Kita asumsikan kode ini sudah benar dari riwayat kita sebelumnya)
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Gagal parse form"})
	}
	title := form.Value["title"]
	artist := form.Value["artist"]
	if len(title) == 0 || len(artist) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Title dan Artist wajib diisi"})
	}
	trackFileHeader := form.File["track_file"][0]
	trackFile, err := trackFileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuka file track"})
	}
	defer trackFile.Close()
	trackFileName := fmt.Sprintf("%v-%s", c.Locals("requestid"), trackFileHeader.Filename)
	trackContentType := trackFileHeader.Header.Get("Content-Type")
	trackOptions := &supabase.FileUploadOptions{
		ContentType: trackContentType,
	}
	trackResp := h.Supabase.Storage.From(h.Config.SUPABASE_BUCKET_tracks).Upload(trackFileName, trackFile, trackOptions)
	if trackResp.Message != "" {
		log.Printf("Error upload track (dari respons): %s", trackResp.Message)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal upload file track", "message": trackResp.Message})
	}
	trackURL := h.Supabase.Storage.From(h.Config.SUPABASE_BUCKET_tracks).GetPublicUrl(trackFileName)
	log.Printf("Berhasil (mencoba) upload file track: %s", trackFileName)

	// ... (kode untuk upload cover)
	coverFileHeader := form.File["cover_file"][0]
	coverFile, err := coverFileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuka file cover"})
	}
	defer coverFile.Close()
	coverExt := filepath.Ext(coverFileHeader.Filename)
	coverFileName := fmt.Sprintf("%v-cover%s", c.Locals("requestid"), coverExt)
	coverContentType := coverFileHeader.Header.Get("Content-Type")
	coverOptions := &supabase.FileUploadOptions{
		ContentType: coverContentType,
	}
	coverResp := h.Supabase.Storage.From(h.Config.SUPABASE_BUCKET_covers).Upload(coverFileName, coverFile, coverOptions)
	if coverResp.Message != "" {
		log.Printf("Error upload cover (dari respons): %s", coverResp.Message)
		h.Supabase.Storage.From(h.Config.SUPABASE_BUCKET_tracks).Remove([]string{trackFileName})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal upload file cover", "message": coverResp.Message})
	}
	coverURL := h.Supabase.Storage.From(h.Config.SUPABASE_BUCKET_covers).GetPublicUrl(coverFileName)
	log.Printf("Berhasil (mencoba) upload file cover: %s", coverFileName)

	// ... (kode untuk simpan ke DB)
	newTrack := Track{
		Title:    title[0],
		Artist:   artist[0],
		URL:      trackURL.SignedUrl,
		PublicID: trackFileName,
		CoverURL: coverURL.SignedUrl,
	}
	if result := h.DB.Create(&newTrack); result.Error != nil {
		// ... (kode rollback Anda)
		h.Supabase.Storage.From(h.Config.SUPABASE_BUCKET_tracks).Remove([]string{trackFileName})
		h.Supabase.Storage.From(h.Config.SUPABASE_BUCKET_covers).Remove([]string{coverFileName})
		log.Printf("Error simpan ke DB: %v\n", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan data track"})
	}
	return c.Status(fiber.StatusCreated).JSON(newTrack)
}


// --- HANDLER BARU 1: Get All Tracks (List Lagu) ---
// (Tetap sama, sudah benar)
func (h *Handler) GetAllTracks(c *fiber.Ctx) error {
	var tracks []Track
	result := h.DB.Order("created_at desc").Find(&tracks)
	if result.Error != nil {
		log.Printf("Error mengambil tracks: %v\n", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data tracks"})
	}
	return c.Status(fiber.StatusOK).JSON(tracks)
}

// --- HANDLER BARU 2: Stream Track (Audio Saja) ---
// (PERBAIKAN: Gunakan Redirect)
func (h *Handler) StreamTrack(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID track tidak valid"})
	}

	var track Track
	result := h.DB.First(&track, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Track tidak ditemukan"})
	}

	if track.URL == "" {
		log.Printf("Gagal streaming track ID %s: URL kosong.", id)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "File audio untuk track ini tidak ditemukan (URL kosong)"})
	}

	// --- ⬇⬇⬇ INI PERBAIKANNYA (Redirect) ⬇⬇⬇ ---
	// (Alih-alih men-download, kita suruh client (Postman)
	//  untuk mengambilnya langsung dari URL Supabase)
	log.Printf("Mengarahkan client ke URL: %s", track.URL)
	return c.Redirect(track.URL, fiber.StatusFound) // StatusFound = 302
	// --- ⬆⬆⬆ AKHIR PERBAIKAN ⬆⬆⬆ ---
}

// --- HANDLER BARU 3: Stream Cover (Gambar Saja) ---
// (PERBAIKAN: Gunakan Redirect)
func (h *Handler) StreamCover(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID track tidak valid"})
	}

	var track Track
	result := h.DB.First(&track, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Track tidak ditemukan"})
	}

	if track.CoverURL == "" {
		log.Printf("Gagal streaming cover ID %s: CoverURL kosong.", id)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "File cover untuk track ini tidak ditemukan (URL kosong)"})
	}

	// --- ⬇⬇⬇ INI PERBAIKANNYA (Redirect) ⬇⬇⬇ ---
	// (Kita suruh client (Postman) mengambilnya langsung)
	log.Printf("Mengarahkan client ke URL: %s", track.CoverURL)
	return c.Redirect(track.CoverURL, fiber.StatusFound) // StatusFound = 302
	// --- ⬆⬆⬆ AKHIR PERBAIKAN ⬆⬆⬆ ---
}

