package track

import (
	"gorm.io/gorm"
)

// Track adalah model database GORM untuk tabel 'tracks'
type Track struct {
	// gorm.Model menambahkan field standar:
	// ID        uint           `gorm:"primarykey"`
	// CreatedAt time.Time
	// UpdatedAt time.Time
	// DeletedAt gorm.DeletedAt `gorm:"index"`
	gorm.Model

	Title    string `json:"title"`
	Artist   string `json:"artist"`
	URL      string `json:"url"`       // URL publik dari Supabase untuk streaming
	PublicID string `json:"public_id"` // Nama file unik di Supabase (untuk hapus/update)
}
