package repository

import (
	"fmt"

	"loan-service/internal/model"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type InvestmentRepository struct{}

func NewInvestmentRepository() *InvestmentRepository {
	return &InvestmentRepository{}
}

func (r *InvestmentRepository) TotalAmountByLoanID(tx *gorm.DB, loanID uint64) (decimal.Decimal, error) {
	var total decimal.Decimal
	if err := tx.
		Model(&model.Investment{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("loan_id = ?", loanID).
		Scan(&total).Error; err != nil {
		return decimal.Zero, fmt.Errorf("sum investments for loan: %w", err)
	}

	return total, nil
}

func (r *InvestmentRepository) Create(tx *gorm.DB, investment model.Investment) error {
	if err := tx.Create(&investment).Error; err != nil {
		return fmt.Errorf("create investment: %w", err)
	}

	return nil
}
