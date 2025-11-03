package track

import (
	"fmt"
	"net/http"

	// "io" <-- Kita tidak jadi pakai ini
	"log"
	"path/filepath"
	"valeth-soundcloud-api/config"

	"github.com/gofiber/fiber/v2"
	// --- PERBAIKAN 1: Import library RESMI yang baru ---
	supabase "github.com/nedpals/supabase-go"
	// --- FIX 1 (BrokenImport): Path gorm yang benar ---
	"gorm.io/gorm"
	// --- Library 'storage_go' yang lama sudah dihapus ---
)

// --- PERBAIKAN 2: Gunakan tipe client yang baru ---
type Handler struct {
	DB       *gorm.DB
	Supabase *supabase.Client // <-- TIPE BARU
	Config   config.Config
}

// --- PERBAIKAN 3: Terima tipe client yang baru ---
func NewHandler(db *gorm.DB, supabase *supabase.Client, config config.Config) *Handler {
	return &Handler{
		DB:       db,
		Supabase: supabase,
		Config:   config,
	}
}

// --- FUNGSI UPLOAD (LOGIKA BARU & FINAL) ---
func (h *Handler) UploadTrack(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Gagal parse form"})
	}

	title := form.Value["title"]
	artist := form.Value["artist"]
	if len(title) == 0 || len(artist) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Title dan Artist wajib diisi"})
	}

	// --- LOGIKA UPLOAD MP3 (PERBAIKAN FINAL) ---
	trackFileHeader := form.File["track_file"][0]
	trackFile, err := trackFileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuka file track"})
	}
	defer trackFile.Close() // <-- Kita defer Close() di sini

	trackFileName := fmt.Sprintf("%v-%s", c.Locals("requestid"), trackFileHeader.Filename)
	trackContentType := trackFileHeader.Header.Get("Content-Type")

	// PANGGILAN UPLOAD (PERBAIKAN FINAL v4)
	// --- FIX 2 (IncompatibleAssign): Hapus '&' ---
	trackOptions := &supabase.FileUploadOptions{
		ContentType: trackContentType, // <-- Hapus '&'
	}
	
	// --- PERBAIKAN v7: TANGKAP RESPONS UPLOAD! ---
	trackResp := h.Supabase.Storage.From(h.Config.SUPABASE_BUCKET_tracks).Upload(trackFileName, trackFile, trackOptions)
	
	// --- LOGGING DEBUG BARU ---
	log.Printf("--- DEBUG UPLOAD TRACK RESPONSE ---")
	log.Printf("Isi Respons: %#v\n", trackResp)
	log.Println("---------------------------------")
	
	// (ASUMSI: Jika 'Message' tidak kosong, berarti error)
	// 'Message' adalah field umum, kita akan coba ini dulu.
	if trackResp.Message != "" {
		log.Printf("Error upload track (dari respons): %s", trackResp.Message)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal upload file track", "message": trackResp.Message})
	}

	// Kita panggil GetPublicUrl secara manual (FIX TYPO: 'U' kecil)
	trackURL := h.Supabase.Storage.From(h.Config.SUPABASE_BUCKET_tracks).GetPublicUrl(trackFileName)

	log.Printf("Berhasil (mencoba) upload file track: %s", trackFileName)

	// --- LOGIKA UPLOAD COVER (PERBAIKAN FINAL) ---
	coverFileHeader := form.File["cover_file"][0]
	coverFile, err := coverFileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuka file cover"})
	}
	defer coverFile.Close() // <-- Kita defer Close() di sini

	coverExt := filepath.Ext(coverFileHeader.Filename)
	coverFileName := fmt.Sprintf("%v-cover%s", c.Locals("requestid"), coverExt)
	coverContentType := coverFileHeader.Header.Get("Content-Type")

	// PANGGILAN UPLOAD (PERBAIKAN FINAL v4)
	// --- FIX 3 (IncompatibleAssign): Hapus '&' ---
	coverOptions := &supabase.FileUploadOptions{
		ContentType: coverContentType, // <-- Hapus '&'
	}

	// --- PERBAIKAN v7: TANGKAP RESPONS UPLOAD! ---
	coverResp := h.Supabase.Storage.From(h.Config.SUPABASE_BUCKET_covers).Upload(coverFileName, coverFile, coverOptions)
	
	// --- LOGGING DEBUG BARU ---
	log.Printf("--- DEBUG UPLOAD COVER RESPONSE ---")
	log.Printf("Isi Respons: %#v\n", coverResp)
	log.Println("---------------------------------")
	
	if coverResp.Message != "" {
		log.Printf("Error upload cover (dari respons): %s", coverResp.Message)
		// Hapus file track yang mungkin sudah berhasil diupload (rollback)
		h.Supabase.Storage.From(h.Config.SUPABASE_BUCKET_tracks).Remove([]string{trackFileName})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal upload file cover", "message": coverResp.Message})
	}

	
	// Kita panggil GetPublicUrl secara manual (FIX TYPO: 'U' kecil)
	coverURL := h.Supabase.Storage.From(h.Config.SUPABASE_BUCKET_covers).GetPublicUrl(coverFileName)
	
	log.Printf("Berhasil (mencoba) upload file cover: %s", coverFileName)


	// 6. Simpan data ke Database PostgreSQL
	newTrack := Track{
		Title:    title[0],
		Artist:   artist[0],
		// --- FIX 4 (MissingField): Ganti 'SignedURL' ke 'SignedUrl' ---
		URL:      trackURL.SignedUrl,
		PublicID: trackFileName,
		// --- FIX 5 (MissingField): Ganti 'SignedURL' ke 'SignedUrl' ---
		CoverURL: coverURL.SignedUrl,
	}

	if result := h.DB.Create(&newTrack); result.Error != nil {
		// ROLLBACK (BARU)
		h.Supabase.Storage.From(h.Config.SUPABASE_BUCKET_tracks).Remove([]string{trackFileName})
		h.Supabase.Storage.From(h.Config.SUPABASE_BUCKET_covers).Remove([]string{coverFileName})

		log.Printf("Error simpan ke DB: %v\n", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan data track"})
	}

	// 7. Kembalikan data track yang baru sebagai JSON
	return c.Status(fiber.StatusCreated).JSON(newTrack)


	
}
// --- ⬇⬇⬇ HANDLER BARU KITA MULAI DARI SINI ⬇⬇⬇ ---

// --- HANDLER BARU 1: Get All Tracks (List Lagu) ---
// (GET /api/v1/tracks)
func (h *Handler) GetAllTracks(c *fiber.Ctx) error {
	var tracks []Track

	// Ambil semua track dari database, urutkan dari yang paling baru
	result := h.DB.Order("created_at desc").Find(&tracks)
	if result.Error != nil {
		log.Printf("Error mengambil tracks: %v\n", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data tracks"})
	}

	// Kembalikan sebagai JSON (lengkap dengan URL track dan cover)
	return c.Status(fiber.StatusOK).JSON(tracks)
}

// --- HANDLER BARU 2: Stream Track (Audio Saja) ---
// (GET /api/v1/tracks/:id/audio)
func (h *Handler) StreamTrack(c *fiber.Ctx) error {
	// 1. Ambil ID dari URL parameter (misal: /api/v1/tracks/1/audio)
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID track tidak valid"})
	}

	// 2. Cari track di database
	var track Track
	result := h.DB.First(&track, id)
	if result.Error != nil {
		// Jika tidak ada track dengan ID itu, kembalikan 404
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Track tidak ditemukan"})
	}

	// 3. Ambil file MP3 dari URL publik Supabase (track.URL)
	// Server Go kita bertindak sebagai 'proxy' untuk men-stream file-nya
	resp, err := http.Get(track.URL)
	if err != nil {
		log.Printf("Error mengambil file dari storage: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil file dari storage"})
	}
	// Pastikan kita menutup body-nya
	defer resp.Body.Close()

	// 4. Set Content-Type header agar browser/Postman tahu ini audio
	c.Set("Content-Type", "audio/mpeg")

	// 5. Stream file-nya (resp.Body) kembali ke client
	return c.SendStream(resp.Body)
}

// --- HANDLER BARU 3: Stream Cover (Gambar Saja) ---
// (GET /api/v1/tracks/:id/cover)
func (h *Handler) StreamCover(c *fiber.Ctx) error {
	// 1. Ambil ID dari URL parameter
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID track tidak valid"})
	}

	// 2. Cari track di database
	var track Track
	result := h.DB.First(&track, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Track tidak ditemukan"})
	}

	// 3. Ambil file Gambar dari URL publik Supabase (track.CoverURL)
	resp, err := http.Get(track.CoverURL)
	if err != nil {
		log.Printf("Error mengambil cover dari storage: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil cover dari storage"})
	}
	defer resp.Body.Close()

	// 4. Set Content-Type (kita tebak saja 'image/png' atau 'image/jpeg')
	//    Browser/Postman cukup pintar untuk menanganinya.
	c.Set("Content-Type", "image/png") // Anda bisa ganti ini jika perlu

	// 5. Stream file-nya kembali ke client
	return c.SendStream(resp.Body)
}
