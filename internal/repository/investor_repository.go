package repository

import (
	"fmt"

	"loan-service/internal/model"

	"gorm.io/gorm"
)

type InvestorRepository struct{}

func NewInvestorRepository() *InvestorRepository {
	return &InvestorRepository{}
}

func (r *InvestorRepository) ExistsByID(tx *gorm.DB, investorID uint64) (bool, error) {
	var count int64
	if err := tx.Model(&model.Investor{}).Where("id = ?", investorID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("check investor exists: %w", err)
	}

	return count > 0, nil
}

func (r *InvestorRepository) GetEmailByID(tx *gorm.DB, investorID uint64) (string, error) {
	var investor model.Investor
	if err := tx.Select("email").First(&investor, investorID).Error; err != nil {
		return "", fmt.Errorf("get investor email: %w", err)
	}

	return investor.Email, nil
}
