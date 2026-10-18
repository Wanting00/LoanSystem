package dto

import (
	"loan-service/internal/model"

	"github.com/shopspring/decimal"
)

type LoanResponse struct {
	ID                 uint64           `json:"id"`
	BorrowerID         uint64           `json:"borrowerId"`
	Principal          decimal.Decimal  `json:"principal"`
	Rate               decimal.Decimal  `json:"rate"`
	ROI                decimal.Decimal  `json:"roi"`
	Status             model.LoanStatus `json:"status"`
	AgreementLetterURL string           `json:"agreementLetterUrl"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type CreateLoanResponse struct {
	ID      uint64 `json:"id"`
	Message string `json:"message"`
}
