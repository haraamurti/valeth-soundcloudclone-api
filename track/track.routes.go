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

}
