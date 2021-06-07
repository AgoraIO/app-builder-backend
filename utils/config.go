// ********************************************
// Copyright © 2021 Agora Lab, Inc., all rights reserved.
// AppBuilder and all associated components, source code, APIs, services, and documentation
// (the “Materials”) are owned by Agora Lab, Inc. and its licensors.  The Materials may not be
// accessed, used, modified, or distributed for any purpose without a license from Agora Lab, Inc.
// Use without a license or in violation of any license terms and conditions (including use for
// any purpose competitive to Agora Lab, Inc.’s business) is strictly prohibited.  For more
// information visit https://appbuilder.agora.io.
// *********************************************

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
	viper.SetDefault("ENABLE_OAUTH", false)
	viper.SetDefault("ENABLE_GOOGLE_OAUTH", false)
	viper.SetDefault("ENABLE_APPLE_OAUTH", false)
	viper.SetDefault("ENABLE_MICROSOFT_OAUTH", false)
	viper.SetDefault("ENABLE_SLACK_OAUTH", false)
	viper.SetDefault("ENABLE_CONSOLE_LOGGING", true)
	viper.SetDefault("ENABLE_FILE_LOGGING", true)
	viper.SetDefault("LOG_LEVEL", "DEBUG")
	viper.SetDefault("ALLOW_LIST", []string{"*"})
	viper.SetDefault("RECORDING_VENDOR", 1)
	viper.SetDefault("RECORDING_REGION", 0)
	viper.SetDefault("RUN_MIGRATION", false)
	viper.SetDefault("PSTN_NUMBER", "(800) 309-2350")

	if viper.GetString("RUN_MIGRATION") == "true" {
		viper.SetDefault("RUN_MIGRATION", true)
	}

	if viper.GetString("ENCRYPTION_ENABLED") == "true" {
		viper.SetDefault("ENCRYPTION_ENABLED", true)
	}

	if viper.GetString("ENABLE_OAUTH") == "false" {
		viper.Set("ENABLE_OAUTH", false)
	}

	if viper.GetString("ENABLE_GOOGLE_OAUTH") == "true" {
		viper.SetDefault("ENABLE_GOOGLE_OAUTH", true)
	}

	if viper.GetString("ENABLE_APPLE_OAUTH") == "true" {
		viper.SetDefault("ENABLE_APPLE_OAUTH", true)
	}

	if viper.GetString("ENABLE_MICROSOFT_OAUTH") == "true" {
		viper.SetDefault("ENABLE_MICROSOFT_OAUTH", true)
	}

	if viper.GetString("ENABLE_SLACK_OAUTH") == "true" {
		viper.SetDefault("ENABLE_SLACK_OAUTH", true)
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

	if viper.GetBool("ENABLE_SLACK_OAUTH") || viper.GetBool("ENABLE_GOOGLE_OAUTH") || viper.GetBool("ENABLE_APPLE_OAUTH") || viper.GetBool("ENABLE_MICROSOFT_OAUTH") {
		viper.SetDefault("ENABLE_OAUTH", true)
	}

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
