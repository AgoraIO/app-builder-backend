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

// SetDefaults sets the default for configuration
func SetDefaults() {
	viper.SetDefault("LOG_DIR", "./logs")
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("MIGRATION_SOURCE", "file://db/migrations") // Will be used in the future
	viper.SetDefault("ALLOWED_ORIGIN", "*")
	viper.SetDefault("ENABLE_OAUTH", true)
	viper.SetDefault("ENABLE_CONSOLE_LOGGING", true)
	viper.SetDefault("ENABLE_FILE_LOGGING", true)
	viper.SetDefault("LOG_LEVEL", "DEBUG")
	viper.SetDefault("ALLOW_LIST", []string{"*"})
	viper.SetDefault("RECORDING_VENDOR", 1)
	viper.SetDefault("RECORDING_REGION", 0)
	viper.SetDefault("RUN_MIGRATION", false)

	if viper.GetString("RUN_MIGRATION") == "true" {
		viper.SetDefault("RUN_MIGRATION", true)
	}

	if viper.GetString("ENCRYPTION_ENABLED") == "true" {
		viper.SetDefault("ENCRYPTION_ENABLED", true)
	}

	if viper.GetString("ENABLE_OAUTH") == "false" {
		viper.Set("ENABLE_OAUTH", false)
	}

	if viper.GetString("ALLOWED_ORIGIN") == "" {
		viper.Set("ALLOWED_ORIGIN", "*")
	}
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

	SetDefaults()

	return CheckRequired()
}

// CheckRequired checks if all the required environment is set
func CheckRequired() error {
	if !viper.IsSet("APP_ID") || !viper.IsSet("APP_CERTIFICATE") || !viper.IsSet("SCHEME") {
		return errors.New("Please Make sure APP_ID,APP_CERTIFICATE and SCHEME are set")
	}

	return nil
}
