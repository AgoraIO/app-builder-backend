package utils

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

// AgoraConfig stores the server side config for token generation
type AgoraConfig struct {
	AppID          string
	AppCertificate string
}

// SetupConfig configures the boilerplate for viper
func SetupConfig(configDir *string) error {
	viper.SetConfigName("config.json")
	viper.SetConfigType("json")

	if configDir == nil {
		viper.AddConfigPath(".")
	} else {
		viper.AddConfigPath(*configDir)
	}

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		return fmt.Errorf("Fatal error config file: %s", err)
	}

	viper.AutomaticEnv()
	return CheckRequired()
}

// CheckRequired checks if all the required environment is set
func CheckRequired() error {
	if !viper.IsSet("APP_ID") || !viper.IsSet("APP_CERTIFICATE") || !viper.IsSet("SCHEME") {
		return errors.New("Please Make sure APP_ID,APP_CERTIFICATE and SCHEME are set")
	}

	return nil
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
