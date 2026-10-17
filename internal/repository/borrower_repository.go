package repository

import (
	"context"
	"fmt"

	"loan-service/internal/model"

	"gorm.io/gorm"
)

type BorrowerRepository struct {
	db *gorm.DB
}

func NewBorrowerRepository(db *gorm.DB) *BorrowerRepository {
	return &BorrowerRepository{db: db}
}

func (r *BorrowerRepository) ExistsByID(ctx context.Context, id uint64) (bool, error) {
	if r.db == nil {
		return false, fmt.Errorf("borrower repository database is nil")
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Borrower{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("check borrower exists: %w", err)
	}

	return count > 0, nil
}
