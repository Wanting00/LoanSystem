package service

import (
	"context"

	"loan-service/internal/dto"
	"loan-service/internal/model"
	"loan-service/internal/repository"
	"loan-service/mapper"

	"github.com/shopspring/decimal"
)

type LoanService struct {
	loanRepository *repository.LoanRepository
}

func NewLoanService(loanRepository *repository.LoanRepository) *LoanService {
	return &LoanService{loanRepository: loanRepository}
}

func (s *LoanService) CreateLoan(ctx context.Context, request dto.CreateLoanRequest) (dto.LoanResponse, error) {
	loan, err := s.loanRepository.Create(ctx, mapper.CreateLoanRequestToLoan(request))
	if err != nil {
		return dto.LoanResponse{}, err
	}

	return mapper.LoanToLoanResponse(loan), nil
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
