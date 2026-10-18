package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Investment struct {
	ID         uint64          `gorm:"primaryKey;autoIncrement"`
	LoanID     uint64          `gorm:"not null;index"`
	InvestorID uint64          `gorm:"not null;index"`
	Amount     decimal.Decimal `gorm:"type:decimal(20,2);not null"`
	CreatedAt  time.Time

	Loan     Loan     `gorm:"foreignKey:LoanID"`
	Investor Investor `gorm:"foreignKey:InvestorID"`
}
