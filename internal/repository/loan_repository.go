package repository

import (
	"context"
	"errors"
	"fmt"

	mysqlDriver "github.com/go-sql-driver/mysql"

	"loan-service/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrDuplicateApproval = errors.New("approval already exists")

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

func (r *LoanRepository) FindByIDForUpdate(tx *gorm.DB, id uint64) (model.Loan, error) {
	var loan model.Loan
	if err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&loan, id).Error; err != nil {
		return model.Loan{}, fmt.Errorf("find loan for update: %w", err)
	}

	return loan, nil
}

func (r *LoanRepository) CreateApproval(tx *gorm.DB, approval model.Approval) error {
	if err := tx.Create(&approval).Error; err != nil {
		var mysqlErr *mysqlDriver.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return fmt.Errorf("%w: %v", ErrDuplicateApproval, err)
		}
		return fmt.Errorf("create approval: %w", err)
	}

	return nil
}

func (r *LoanRepository) UpdateLoanStatus(tx *gorm.DB, id uint64, status model.LoanStatus) error {
	result := tx.
		Model(&model.Loan{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("update loan status: %w", result.Error)
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("update loan status: expected one affected row, got %d", result.RowsAffected)
	}

	return nil
}
