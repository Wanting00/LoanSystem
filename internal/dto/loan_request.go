package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreateLoanRequest struct {
	BorrowerID         uint64           `json:"borrowerId" binding:"required"`
	Principal          *decimal.Decimal `json:"principal" binding:"required"`
	Rate               *decimal.Decimal `json:"rate" binding:"required"`
	ROI                *decimal.Decimal `json:"roi" binding:"required"`
	AgreementLetterURL string           `json:"agreementLetterUrl" binding:"required"`
}

type ApproveLoanRequest struct {
	ValidatorPhotoURL string    `json:"validatorPhotoUrl" binding:"required"`
	EmployeeID        string    `json:"employeeId" binding:"required"`
	ApprovalDate      time.Time `json:"approvalDate" binding:"required"`
}

type InvestRequest struct {
	InvestorID uint64           `json:"investorId" binding:"required"`
	Amount     *decimal.Decimal `json:"amount" binding:"required"`
}

type DisburseRequest struct {
	SignedAgreementURL string    `json:"signedAgreementUrl"`
	EmployeeID         string    `json:"employeeId"`
	DisbursementDate   time.Time `json:"disbursementDate"`
}
