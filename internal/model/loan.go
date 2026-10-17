package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type LoanStatus string

const (
	LoanStatusProposed  LoanStatus = "PROPOSED"
	LoanStatusApproved  LoanStatus = "APPROVED"
	LoanStatusInvested  LoanStatus = "INVESTED"
	LoanStatusDisbursed LoanStatus = "DISBURSED"
)

type Loan struct {
	ID                 uint64          `gorm:"primaryKey;autoIncrement"`
	BorrowerID         uint64          `gorm:"not null;index"`
	Principal          decimal.Decimal `gorm:"type:decimal(20,2);not null"`
	Rate               decimal.Decimal `gorm:"type:decimal(10,6)"`
	ROI                decimal.Decimal `gorm:"type:decimal(10,6)"`
	AgreementLetterURL string          `gorm:"type:varchar(2048)"`
	Status             LoanStatus      `gorm:"not null;index"`
	CreatedAt          time.Time
	UpdatedAt          time.Time

	Borrower     *Borrower     `gorm:"foreignKey:BorrowerID"`
	Approval     *Approval     `gorm:"foreignKey:LoanID"`
	Investments  []Investment  `gorm:"foreignKey:LoanID"`
	Disbursement *Disbursement `gorm:"foreignKey:LoanID"`
}
