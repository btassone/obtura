package config

import (
	"os"
	"strconv"
	"time"

	"github.com/btassone/obtura/pkg/database"
)

// GetDatabaseConfig returns database configuration based on environment
func GetDatabaseConfig() *database.Config {
	// Check if we're in development mode
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	// Default to SQLite for development
	if env == "development" {
		return &database.Config{
			Driver:          "sqlite3",
			SQLitePath:      getEnv("DB_PATH", "./data/obtura.db"),
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
		}
	}

	// For production, use configured database
	driver := getEnv("DB_DRIVER", "postgres")
	
	config := &database.Config{
		Driver:          driver,
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            getEnvAsInt("DB_PORT", getDefaultPort(driver)),
		Database:        getEnv("DB_NAME", "obtura"),
		Username:        getEnv("DB_USER", "obtura"),
		Password:        getEnv("DB_PASS", ""),
		MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: time.Duration(getEnvAsInt("DB_CONN_MAX_LIFETIME_MIN", 5)) * time.Minute,
	}

	return config
}

// getDefaultPort returns the default port for a database driver
func getDefaultPort(driver string) int {
	switch driver {
	case "mysql":
		return 3306
	case "postgres", "postgresql":
		return 5432
	default:
		return 5432
	}
}

// getEnv gets an environment variable with a fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvAsInt gets an environment variable as an integer with a fallback
func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}