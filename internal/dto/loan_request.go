package dto

import "time"

type CreateLoanRequest struct {
	BorrowerID         uint64  `json:"borrowerId"`
	Principal          int64   `json:"principal"`
	Rate               float64 `json:"rate"`
	ROI                float64 `json:"roi"`
	AgreementLetterURL string  `json:"agreementLetterUrl"`
}

type ApproveLoanRequest struct {
	ValidatorPhotoURL string    `json:"validatorPhotoUrl"`
	EmployeeID        string    `json:"employeeId"`
	ApprovalDate      time.Time `json:"approvalDate"`
}

type InvestRequest struct {
	InvestorID uint64 `json:"investorId"`
	Amount     int64  `json:"amount"`
}

type DisburseRequest struct {
	SignedAgreementURL string    `json:"signedAgreementUrl"`
	EmployeeID         string    `json:"employeeId"`
	DisbursementDate   time.Time `json:"disbursementDate"`
}
