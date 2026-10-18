package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"

	"loan-service/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrDuplicateApproval = errors.New("approval already exists")
var ErrDuplicateIdempotencyKey = errors.New("loan idempotency key already exists")

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

	return r.CreateInTx(r.db.WithContext(ctx), loan)
}

func (r *LoanRepository) CreateInTx(tx *gorm.DB, loan model.Loan) (model.Loan, error) {
	if tx == nil {
		return model.Loan{}, fmt.Errorf("create loan: database transaction is nil")
	}

	if err := tx.Create(&loan).Error; err != nil {
		var mysqlErr *mysqlDriver.MySQLError
		if errors.As(err, &mysqlErr) &&
			mysqlErr.Number == 1062 &&
			isIdempotencyKeyIndexConflict(mysqlErr.Message) {
			return model.Loan{}, fmt.Errorf("%w: %v", ErrDuplicateIdempotencyKey, err)
		}
		return model.Loan{}, fmt.Errorf("create loan: %w", err)
	}

	return loan, nil
}

func isIdempotencyKeyIndexConflict(message string) bool {
	const keyMarker = "for key "
	keyMarkerIndex := strings.LastIndex(message, keyMarker)
	if keyMarkerIndex == -1 {
		return false
	}

	keyName := strings.Trim(strings.TrimSpace(message[keyMarkerIndex+len(keyMarker):]), "'`\"")
	if tableSeparator := strings.LastIndex(keyName, "."); tableSeparator != -1 {
		keyName = keyName[tableSeparator+1:]
	}

	return keyName == "ux_loans_idempotency_key"
}

func (r *LoanRepository) FindByIdempotencyKey(db *gorm.DB, key string) (model.Loan, error) {
	if db == nil {
		return model.Loan{}, fmt.Errorf("find loan by idempotency key: database connection is nil")
	}

	var loan model.Loan
	if err := db.Where("idempotency_key = ?", key).First(&loan).Error; err != nil {
		return model.Loan{}, fmt.Errorf("find loan by idempotency key: %w", err)
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
