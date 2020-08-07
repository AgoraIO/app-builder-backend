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

// SetupConfig configures the boilerplate for viper
func SetupConfig() {
	viper.SetConfigName("config.json")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}

	viper.AutomaticEnv()
}

// GetPORT fetches the PORT
func GetPORT() string {
	if port := viper.GetString("PORT"); port != "" {
		return port
	}

	return "8080"
}

// GetDBURL fetches the database string
func GetDBURL() string {
	return viper.GetString("DATABASE_URL")
}

// GetAgoraConfig returns an AgoraConfig based on what's present in the envfile
func GetAgoraConfig() AgoraConfig {
	return AgoraConfig{
		AppID:          viper.GetString("appID"),
		AppCertificate: viper.GetString("appCertificate"),
	}
}
