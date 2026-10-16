package model

import "time"

type LoanStatus string

const (
	LoanStatusProposed  LoanStatus = "PROPOSED"
	LoanStatusApproved  LoanStatus = "APPROVED"
	LoanStatusInvested  LoanStatus = "INVESTED"
	LoanStatusDisbursed LoanStatus = "DISBURSED"
)

type Loan struct {
	ID                 uint64  `gorm:"primaryKey;autoIncrement"`
	BorrowerID         uint64  `gorm:"not null;index"`
	Principal          int64   `gorm:"not null"`
	Rate               float64 `gorm:"not null"`
	ROI                float64 `gorm:"not null"`
	AgreementLetterURL string
	Status             LoanStatus `gorm:"not null;index"`
	CreatedAt          time.Time
	UpdatedAt          time.Time

	Borrower     Borrower      `gorm:"foreignKey:BorrowerID"`
	Approval     *Approval     `gorm:"foreignKey:LoanID"`
	Investments  []Investment  `gorm:"foreignKey:LoanID"`
	Disbursement *Disbursement `gorm:"foreignKey:LoanID"`
}
