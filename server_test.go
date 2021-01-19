package main

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

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
}

func GetString(key string) string {
	return viper.GetString(key)
}

func TestConfig(t *testing.T) {
	SetupConfig()
	configList := []string{
		"APP_ID",
		"APP_CERTIFICATE",
		"CUSTOMER_ID",
		"CUSTOMER_CERTIFICATE",
		"BUCKET_NAME",
		"BUCKET_ACCESS_KEY",
		"BUCKET_ACCESS_SECRET",
		"DATABASE_URL",
	}
	if GetString("ENABLE_OAUTH") == "true" || viper.GetBool("ENABLE_OAUTH") {
		configList = append(configList, "CLIENT_ID", "CLIENT_SECRET")
	}
	for _, key := range configList {
		assert.NotEqual(t, "", GetString(key), fmt.Sprintf("Error in %v", key))
		assert.Equal(t, "string", reflect.TypeOf(GetString(key)).String(), fmt.Sprintf("%v should be string", key))
	}
}
