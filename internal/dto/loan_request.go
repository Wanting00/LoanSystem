package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreateLoanRequest struct {
	BorrowerID         uint64          `json:"borrowerId"`
	Principal          decimal.Decimal `json:"principal"`
	Rate               decimal.Decimal `json:"rate"`
	ROI                decimal.Decimal `json:"roi"`
	AgreementLetterURL string          `json:"agreementLetterUrl"`
}

type ApproveLoanRequest struct {
	ValidatorPhotoURL string    `json:"validatorPhotoUrl" binding:"required"`
	EmployeeID        string    `json:"employeeId" binding:"required"`
	ApprovalDate      time.Time `json:"approvalDate" binding:"required"`
}

type InvestRequest struct {
	InvestorID uint64          `json:"investorId"`
	Amount     decimal.Decimal `json:"amount"`
}

type DisburseRequest struct {
	SignedAgreementURL string    `json:"signedAgreementUrl"`
	EmployeeID         string    `json:"employeeId"`
	DisbursementDate   time.Time `json:"disbursementDate"`
}
