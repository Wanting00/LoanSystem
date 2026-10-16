package model

import "time"

type Investment struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement"`
	LoanID     uint64 `gorm:"not null;index"`
	InvestorID uint64 `gorm:"not null;index"`
	Amount     int64  `gorm:"not null"`
	CreatedAt  time.Time

	Loan     Loan     `gorm:"foreignKey:LoanID"`
	Investor Investor `gorm:"foreignKey:InvestorID"`
}
