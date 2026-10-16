package dto

import "loan-service/internal/model"

type LoanResponse struct {
	ID                 uint64           `json:"id"`
	BorrowerID         uint64           `json:"borrowerId"`
	Principal          int64            `json:"principal"`
	Rate               float64          `json:"rate"`
	ROI                float64          `json:"roi"`
	Status             model.LoanStatus `json:"status"`
	AgreementLetterURL string           `json:"agreementLetterUrl"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
