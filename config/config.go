package config

import (
	"errors"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Config struct {
	HTTPAddress string
	MySQLDSN    string
}

func Load() Config {
	return Config{
		HTTPAddress: envOrDefault("HTTP_ADDR", ":8080"),
		MySQLDSN:    os.Getenv("MYSQL_DSN"),
	}
}

func OpenMySQL(cfg Config) (*gorm.DB, error) {
	if cfg.MySQLDSN == "" {
		return nil, errors.New("MYSQL_DSN must be set before opening the database")
	}

	return gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
