package config

import "os"

type Config struct {
	MongoDB MongoDBConfig
}

type MongoDBConfig struct {
	URI      string
	Database string
}

func LoadConfig() *Config {
	return &Config{
		MongoDB: MongoDBConfig{
			URI:      getEnvOrDefault("MONGODB_URI", "mongodb://admin:password123@localhost:27017/?authSource=admin"),
			Database: getEnvOrDefault("MONGODB_DATABASE", "auth_db"),
		},
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
