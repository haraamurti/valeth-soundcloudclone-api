package config

import (
	"fmt"

	"github.com/spf13/viper"
)



type Config struct {
	DatabaseURL    string `mapstructure:"db_URL"` //menset database url untuk dari supabase
	SupabaseURL    string `mapstructure:"SUPABASE_URL"` // akan dipakai nanti
	SupabaseKey    string `mapstructure:"SUPABASE_KEY"`//akan dipakai nanti
	SupabaseBucket string `mapstructure:"SUPABASE_BUCKET"`//akan dipakai nanti
}


func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)   //
	viper.SetConfigName(".env") // The name of the config file (without extension)
	viper.SetConfigType("env")  // The type of the file

	//viper.AutomaticEnv() // this is going to read the env from the os enviorment. if we have to

	// this will read the config file.
	err = viper.ReadInConfig()
	if err != nil {
		return
		fmt.Println("erorr reading config..")
	}

	// "Unmarshal" the loaded values into our Config struct
	err = viper.Unmarshal(&config)
	return
}

