package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort string
	LogLevel   string
	JWTSecret  string
	DBPath     string
}

func LoadConfig() *Config {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key" // В продакшене нужно использовать безопасный ключ
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "calc.db"
	}

	return &Config{
		ServerPort: port,
		LogLevel:   logLevel,
		JWTSecret:  jwtSecret,
		DBPath:     dbPath,
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
