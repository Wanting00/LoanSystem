package service

import (
	"loan-service/internal/dto"
	"loan-service/internal/model"

	"github.com/shopspring/decimal"
)

type LoanService struct{}

func NewLoanService() *LoanService {
	return &LoanService{}
}

func (s *LoanService) CreateLoan(request dto.CreateLoanRequest) dto.LoanResponse {
	return dto.LoanResponse{
		ID:                 1,
		BorrowerID:         request.BorrowerID,
		Principal:          request.Principal,
		Rate:               request.Rate,
		ROI:                request.ROI,
		Status:             model.LoanStatusProposed,
		AgreementLetterURL: request.AgreementLetterURL,
	}
}

func (s *LoanService) GetLoan(id string) dto.LoanResponse {
	return dto.LoanResponse{
		ID:                 1,
		BorrowerID:         1,
		Principal:          decimal.NewFromInt(100000),
		Rate:               decimal.NewFromFloat(0.05),
		ROI:                decimal.NewFromFloat(0.08),
		Status:             model.LoanStatusProposed,
		AgreementLetterURL: "https://example.com/agreement.pdf",
	}
}

func (s *LoanService) ApproveLoan(id string, request dto.ApproveLoanRequest) dto.MessageResponse {
	return mockResponse()
}

func (s *LoanService) InvestInLoan(id string, request dto.InvestRequest) dto.MessageResponse {
	return mockResponse()
}

func (s *LoanService) DisburseLoan(id string, request dto.DisburseRequest) dto.MessageResponse {
	return mockResponse()
}

func mockResponse() dto.MessageResponse {
	return dto.MessageResponse{Message: "mock response"}
}
