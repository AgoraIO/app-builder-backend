package utils

import (
	"fmt"

	"github.com/spf13/viper"
)

// AgoraConfig stores the server side config for token generation
type AgoraConfig struct {
	AppID          string
	AppCertificate string
}

// SetDefaults sets the default for configuration
func SetDefaults() {
	viper.SetDefault("LOG_DIR", "./logs")
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("MIGRATION_SOURCE", "file://db/migrations") // Will be used later
	viper.SetDefault("ALLOWED_ORIGIN", "*")
	viper.SetDefault("ENABLE_OAUTH", true)
}

// SetupConfig configures the boilerplate for viper
func SetupConfig() {
	viper.SetConfigName("config.json")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}

	SetDefaults()
	viper.AutomaticEnv()
}
