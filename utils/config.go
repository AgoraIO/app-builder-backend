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

// GetConfig returns an AgoraConfig based on what's present in the envfile
func GetConfig() AgoraConfig {
	viper.SetConfigName("envfile")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}

	return AgoraConfig{
		AppID:          viper.GetString("appID"),
		AppCertificate: viper.GetString("appCertificate"),
	}
}
