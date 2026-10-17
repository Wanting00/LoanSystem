package repository

import (
	"context"
	"fmt"

	"loan-service/internal/model"

	"gorm.io/gorm"
)

type LoanRepository struct {
	db *gorm.DB
}

func NewLoanRepository(db *gorm.DB) *LoanRepository {
	return &LoanRepository{db: db}
}

func (r *LoanRepository) Create(ctx context.Context, loan model.Loan) (model.Loan, error) {
	if r.db == nil {
		return model.Loan{}, fmt.Errorf("loan repository database is nil")
	}

	if err := r.db.WithContext(ctx).Create(&loan).Error; err != nil {
		return model.Loan{}, fmt.Errorf("create loan: %w", err)
	}

	return loan, nil
}
