package config

import (
	"os"
	"strconv"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	KubeconfigPath   string
	Namespace        string
	StorageClass     string
	Port             int
	LogLevel         string
	MaxConcurrentVMs int
}

// Load reads configuration from environment variables and config file
func Load() *Config {
	// Set defaults
	viper.SetDefault("kubeconfig", getEnvOrDefault("KUBECONFIG", ""))
	viper.SetDefault("namespace", getEnvOrDefault("NAMESPACE", "goldenpipe-system"))
	viper.SetDefault("storage-class", getEnvOrDefault("STORAGE_CLASS", "rook-ceph-block"))
	viper.SetDefault("port", getEnvIntOrDefault("API_PORT", 8080))
	viper.SetDefault("log-level", getEnvOrDefault("LOG_LEVEL", "info"))
	viper.SetDefault("max-concurrent-vms", getEnvIntOrDefault("MAX_CONCURRENT_VMS", 5))

	return &Config{
		KubeconfigPath:   viper.GetString("kubeconfig"),
		Namespace:        viper.GetString("namespace"),
		StorageClass:     viper.GetString("storage-class"),
		Port:             viper.GetInt("port"),
		LogLevel:         viper.GetString("log-level"),
		MaxConcurrentVMs: viper.GetInt("max-concurrent-vms"),
	}
}

// getEnvOrDefault returns the value of an environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault returns the integer value of an environment variable or a default value
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
