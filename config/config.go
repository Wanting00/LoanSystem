package config

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Config struct {
	HTTPAddress   string
	MySQLHost     string
	MySQLPort     string
	MySQLDatabase string
	MySQLUser     string
	MySQLPassword string
}

func Load() Config {
	return Config{
		HTTPAddress:   envOrDefault("HTTP_ADDR", ":8080"),
		MySQLHost:     envOrDefault("MYSQL_HOST", "127.0.0.1"),
		MySQLPort:     envOrDefault("MYSQL_PORT", "3306"),
		MySQLDatabase: envOrDefault("MYSQL_DATABASE", "loan_system"),
		MySQLUser:     os.Getenv("MYSQL_USER"),
		MySQLPassword: os.Getenv("MYSQL_PASSWORD"),
	}
}

func OpenMySQL(cfg Config) (*gorm.DB, error) {
	if cfg.MySQLUser == "" {
		return nil, fmt.Errorf("MYSQL_USER must be set before opening the database")
	}

	dsn, err := mysqlDSN(cfg, cfg.MySQLDatabase)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connect to MySQL at %s:%s database %q: %w", cfg.MySQLHost, cfg.MySQLPort, cfg.MySQLDatabase, err)
	}

	return db, nil
}

func EnsureMySQLDatabase(cfg Config) error {
	dsn, err := mysqlDSN(cfg, "")
	if err != nil {
		return err
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("connect to MySQL server at %s:%s: %w", cfg.MySQLHost, cfg.MySQLPort, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get MySQL connection pool: %w", err)
	}
	defer sqlDB.Close()

	database := "`" + strings.ReplaceAll(cfg.MySQLDatabase, "`", "``") + "`"
	if err := db.Exec("CREATE DATABASE IF NOT EXISTS " + database + " CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci").Error; err != nil {
		return fmt.Errorf("create MySQL database %q: %w", cfg.MySQLDatabase, err)
	}

	return nil
}

func mysqlDSN(cfg Config, database string) (string, error) {
	if cfg.MySQLUser == "" {
		return "", fmt.Errorf("MYSQL_USER must be set before connecting to MySQL")
	}
	if cfg.MySQLPassword == "" {
		return "", fmt.Errorf("MYSQL_PASSWORD must be set before connecting to MySQL")
	}
	if cfg.MySQLHost == "" || cfg.MySQLPort == "" {
		return "", fmt.Errorf("MYSQL_HOST and MYSQL_PORT must be set before connecting to MySQL")
	}
	if database != "" && database != cfg.MySQLDatabase {
		return "", fmt.Errorf("database name does not match configured MYSQL_DATABASE")
	}
	if cfg.MySQLDatabase == "" {
		return "", fmt.Errorf("MYSQL_DATABASE must be set before connecting to MySQL")
	}

	return (&mysqlDriver.Config{
		User:      cfg.MySQLUser,
		Passwd:    cfg.MySQLPassword,
		Net:       "tcp",
		Addr:      net.JoinHostPort(cfg.MySQLHost, cfg.MySQLPort),
		DBName:    database,
		ParseTime: true,
		Loc:       time.Local,
		Params:    map[string]string{"charset": "utf8mb4"},
	}).FormatDSN(), nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
