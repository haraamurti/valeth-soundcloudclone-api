package storage

import (
	"log"
	"valeth-soundcloud-api/config"

	storage_go "github.com/supabase-community/storage-go"
)

func InitSupabase(config config.Config) *storage_go.Client /*storage_go.Client this is a return value*/ {
	
	//making variable to connect
	client := storage_go.NewClient(config.SupabaseURL, config.SupabaseKey, nil)

	log.Println("Supabase Storage client initialized.")
	return client
}

