package model

import "time"

type Disbursement struct {
	ID                 uint64    `gorm:"primaryKey;autoIncrement"`
	LoanID             uint64    `gorm:"not null;uniqueIndex"`
	SignedAgreementURL string    `gorm:"not null"`
	EmployeeID         string    `gorm:"not null"`
	DisbursementDate   time.Time `gorm:"not null"`
	CreatedAt          time.Time

	Loan Loan `gorm:"foreignKey:LoanID"`
}
