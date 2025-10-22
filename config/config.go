package config

import (
	"github.com/spf13/viper"
)

// Config struct holds all configuration for the application.
// The `mapstructure` tags tell Viper which .env key maps to which struct field.
type Config struct {
	DatabaseURL    string `mapstructure:"db_URL"`
	SupabaseURL    string `mapstructure:"SUPABASE_URL"`
	SupabaseKey    string `mapstructure:"SUPABASE_KEY"`
	SupabaseBucket string `mapstructure:"SUPABASE_BUCKET"`
}

// LoadConfig reads configuration from the .env file in the specified path.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)   // Tell viper where to look for the file
	viper.SetConfigName(".env") // The name of the config file (without extension)
	viper.SetConfigType("env")  // The type of the file

	viper.AutomaticEnv() // Also read from system environment variables if available

	// Find and read the config file
	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	// "Unmarshal" the loaded values into our Config struct
	err = viper.Unmarshal(&config)
	return
}

