package main

import (
	"fmt"
	"log"

	"loan-service/config"
	"loan-service/migrations"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.Load()
	if err := config.EnsureMySQLDatabase(cfg); err != nil {
		return err
	}

	db, err := config.OpenMySQL(cfg)
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get MySQL connection pool: %w", err)
	}
	defer sqlDB.Close()

	if err := migrations.AutoMigrate(db); err != nil {
		return err
	}

	log.Printf("database %q created and models migrated successfully", cfg.MySQLDatabase)
	return nil
}
