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
	viper.SetDefault("MIGRATION_SOURCE", "file://db/migrations") // Will be used in the future
	viper.SetDefault("ALLOWED_ORIGIN", "*")
	viper.SetDefault("ENABLE_OAUTH", true)
	viper.SetDefault("ENABLE_CONSOLE_LOGGINIG", true)
	viper.SetDefault("ENABLE_FILE_LOGGING", true)
	viper.SetDefault("LOG_LEVEL", "DEBUG")

	if viper.GetString("ALLOWED_ORIGIN") == "" {
		viper.Set("ALLOWED_ORIGIN", "*")
	}
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

	viper.SetDefault("RECORDING_VENDOR", 1)
	viper.SetDefault("RECORDING_REGION", 0)
	viper.AutomaticEnv()
	if viper.GetString("ENABLE_OAUTH") == "false" {
		viper.Set("ENABLE_OAUTH", false)
	}

	SetDefaults()
}

// GetPORT fetches the PORT
func GetPORT(defaultPort string) string {
	if port := viper.GetString("PORT"); port != "" {
		return port
	}

	return defaultPort
}

// GetDBURL fetches the database string
func GetDBURL() string {
	return viper.GetString("DATABASE_URL")
}

// GetMigrationSource gets the url from which the database migrations are fetched from
func GetMigrationSource() string {
	if source := viper.GetString("MIGRATION_SOURCE"); source != "" {
		return source
	}

	return "file://db/migrations"
}

// GetAllowedOrigin returns origin in the environment variable, else allows all origins
func GetAllowedOrigin() string {
	if origin := viper.GetString("ALLOWED_ORIGIN"); origin != "" {
		return origin
	}

	return "*"
}
