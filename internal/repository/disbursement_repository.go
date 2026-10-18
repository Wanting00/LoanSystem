package repository

import (
	"errors"
	"fmt"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"loan-service/internal/model"
)

var ErrDuplicateDisbursement = errors.New("disbursement already exists")

type DisbursementRepository struct{}

func NewDisbursementRepository() *DisbursementRepository {
	return &DisbursementRepository{}
}

func (r *DisbursementRepository) Create(tx *gorm.DB, disbursement *model.Disbursement) error {
	if err := tx.Create(disbursement).Error; err != nil {
		var mysqlErr *mysqlDriver.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrDuplicateDisbursement
		}
		return fmt.Errorf("create disbursement: %w", err)
	}

	return nil
}
