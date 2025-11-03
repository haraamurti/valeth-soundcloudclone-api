package track

import (
	"github.com/gofiber/fiber/v2"
)

//akan setup track routes untuk handler ini
func SetupTrackRoutes(api fiber.Router, handler *Handler) {
	//kita akan membuat group dari endpoint tracks ini
	tracks := api.Group("/tracks")

	// Daftarkan endpoint POST /api/v1/tracks/upload
	//link unutk upload.
	tracks.Post("/upload", handler.UploadTrack)
	tracks.Get("/u", handler.UploadTrack)


	// --- ⬇⬇⬇ RUTE-RUTE BARU KITA ⬇⬇⬇ ---

	// RUTE 1: GET /api/v1/tracks
	// (Mengambil daftar semua lagu sebagai JSON)
	tracks.Get("/", handler.GetAllTracks)

	// RUTE 2: GET /api/v1/tracks/1/audio
	// (Memainkan/streaming audio dari track dengan ID 1)
	tracks.Get("/:id/audio", handler.StreamTrack)

	// RUTE 3: GET /api/v1/tracks/1/cover
	// (Melihat/streaming gambar cover dari track dengan ID 1)
	tracks.Get("/:id/cover", handler.StreamCover)

}
