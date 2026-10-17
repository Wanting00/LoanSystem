package migrations

import (
	"fmt"

	"loan-service/internal/model"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	if err := db.AutoMigrate(
		&model.Borrower{},
		&model.Investor{},
		&model.Loan{},
		&model.Approval{},
		&model.Investment{},
		&model.Disbursement{},
	); err != nil {
		return fmt.Errorf("auto-migrate loan system models: %w", err)
	}

	return nil
}
