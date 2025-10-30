package storage

import (
	"fmt" // <-- IMPORT BARU
	"log"
	"valeth-soundcloud-api/config"

	// <-- IMPORT BARU YANG SUDAH BENAR
	supabase "github.com/nedpals/supabase-go"
)

// InitSupabase (BARU) - perhatikan sekarang dia mengembalikan (Client, error)
// Ini adalah "pabrik" untuk client resmi Supabase
func InitSupabase(config config.Config) (*supabase.Client, error) {
	// Library resmi punya cara inisialisasi yang berbeda
	// PERBAIKAN: CreateClient HANYA mengembalikan *supabase.Client, tidak mengembalikan error.
	client := supabase.CreateClient(config.SupabaseURL, config.SupabaseKey)

	// Kita tidak bisa mengecek error saat inisialisasi.
	// Kita anggap jika URL/Key salah, error akan terjadi saat request pertama (misal, upload).
	// Kita bisa tambahkan cek nil sederhana sebagai penjagaan.
	if client == nil {
		err := fmt.Errorf("gagal membuat client supabase, client adalah nil")
		log.Println(err)
		return nil, err
	}

	log.Println("Supabase client (RESMI) berhasil diinisialisasi.")
	// Kita tetap kembalikan (client, nil) agar sesuai dengan signature baru kita,
	// yang akan ditangani oleh main.go.
	return client, nil
}

