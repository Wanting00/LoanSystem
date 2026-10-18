package service

import (
	"context"
	"errors"
	"strings"

	"loan-service/internal/dto"
	"loan-service/internal/model"
	"loan-service/internal/repository"
	"loan-service/mapper"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

var (
	ErrLoanNotFound           = errors.New("loan not found")
	ErrLoanNotProposed        = errors.New("loan is not in proposed status")
	ErrApprovalConflict       = errors.New("approval record already exists")
	ErrInvalidApprovalRequest = errors.New("approval information is required")
)

type LoanService struct {
	db              *gorm.DB
	loanRepository  *repository.LoanRepository
	borrowerService *BorrowerService
}

func NewLoanService(db *gorm.DB, loanRepository *repository.LoanRepository, borrowerService *BorrowerService) *LoanService {
	return &LoanService{db: db, loanRepository: loanRepository, borrowerService: borrowerService}
}

func (s *LoanService) CreateLoan(ctx context.Context, request dto.CreateLoanRequest) (uint64, error) {
	// Check if the borrower exists before creating the loan
	exists, err := s.borrowerService.BorrowerExists(ctx, request.BorrowerID)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, ErrBorrowerNotFound
	}

	//TODO: idempotency check before creating it

	// TODO: check if the loan terms are valid before creating the loan

	// Create the loan in the repository
	loan, err := s.loanRepository.Create(ctx, mapper.CreateLoanRequestToLoan(request))
	if err != nil {
		return 0, err
	}

	return loan.ID, nil
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

func (s *LoanService) ApproveLoan(ctx context.Context, id uint64, request dto.ApproveLoanRequest) error {
	if strings.TrimSpace(request.ValidatorPhotoURL) == "" ||
		strings.TrimSpace(request.EmployeeID) == "" ||
		request.ApprovalDate.IsZero() {
		return ErrInvalidApprovalRequest
	}

	// The transaction commits if the callback returns nil
	// rolls back on any returned error.
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.approveLoanInTx(tx, id, request)
	})
}

// approveLoanInTx runs every approval step on tx so they commit or roll back together.
func (s *LoanService) approveLoanInTx(tx *gorm.DB, id uint64, request dto.ApproveLoanRequest) error {
	loan, err := s.loanRepository.FindByIDForUpdate(tx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrLoanNotFound
	}
	if err != nil {
		return err
	}
	if loan.Status != model.LoanStatusProposed {
		return ErrLoanNotProposed
	}

	err = s.loanRepository.CreateApproval(tx, mapper.ApproveLoanRequestToApproval(loan.ID, request))
	if errors.Is(err, repository.ErrDuplicateApproval) {
		return ErrApprovalConflict
	}
	if err != nil {
		return err
	}

	return s.loanRepository.UpdateLoanStatus(tx, loan.ID, model.LoanStatusApproved)
}

func (s *LoanService) InvestInLoan(id string, request dto.InvestRequest) dto.MessageResponse {
	return mockResponse()
}

func mockResponse() dto.MessageResponse {
	return dto.MessageResponse{Message: "mock response"}
}
