package model

import "time"

type Approval struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement"`
	LoanID            uint64    `gorm:"not null;uniqueIndex"`
	ValidatorPhotoURL string    `gorm:"not null"`
	EmployeeID        string    `gorm:"not null"`
	ApprovalDate      time.Time `gorm:"not null"`
	CreatedAt         time.Time

	Loan Loan `gorm:"foreignKey:LoanID"`
}
