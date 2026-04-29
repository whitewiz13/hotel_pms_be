package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Upload   UploadConfig
}

type ServerConfig struct {
	Port    string
	GinMode string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	Timezone string
}

type JWTConfig struct {
	Secret           string
	ExpiryHours      int
	GuestExpiryHours int
}

type UploadConfig struct {
	Dir     string // local directory for file uploads
	BaseURL string // public base URL for serving uploaded files
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		d.Host, d.User, d.Password, d.DBName, d.Port, d.SSLMode, d.Timezone,
	)
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	jwtExpiry, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	guestExpiry, _ := strconv.Atoi(getEnv("JWT_GUEST_EXPIRY_HOURS", "12"))

	cfg := &Config{
		Server: ServerConfig{
			Port:    getEnv("SERVER_PORT", "8080"),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "hotel_pms"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			Timezone: getEnv("DB_TIMEZONE", "UTC"),
		},
		JWT: JWTConfig{
			Secret:           getEnv("JWT_SECRET", ""),
			ExpiryHours:      jwtExpiry,
			GuestExpiryHours: guestExpiry,
		},
		Upload: UploadConfig{
			Dir:     getEnv("UPLOAD_DIR", "./uploads"),
			BaseURL: getEnv("UPLOAD_BASE_URL", "http://localhost:8080"),
		},
	}

	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func ShouldRunMigrations() bool {
	return getEnv("RUN_MIGRATIONS", "true") == "true"
}

func ShouldSeedDB() bool {
	return getEnv("SEED_DB", "false") == "true"
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
