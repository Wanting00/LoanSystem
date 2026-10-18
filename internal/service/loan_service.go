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
	ErrIdempotencyConflict    = errors.New("idempotency key was already used with different loan parameters")
)

type LoanService struct {
	db              *gorm.DB
	loanRepository  *repository.LoanRepository
	borrowerService *BorrowerService
}

func NewLoanService(
	db *gorm.DB,
	loanRepository *repository.LoanRepository,
	borrowerService *BorrowerService) *LoanService {
	return &LoanService{db: db, loanRepository: loanRepository, borrowerService: borrowerService}
}

func (s *LoanService) CreateLoan(ctx context.Context, idempotencyKey string, request dto.CreateLoanRequest) (uint64, error) {
	existingLoan, err := s.loanRepository.FindByIdempotencyKey(s.db.WithContext(ctx), idempotencyKey)
	if err == nil {
		return matchingLoanID(existingLoan, request)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	// Check if the borrower exists before creating the loan
	exists, err := s.borrowerService.BorrowerExists(ctx, request.BorrowerID)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, ErrBorrowerNotFound
	}

	// TODO: check if the loan terms are valid before creating the loan

	loan := mapper.CreateLoanRequestToLoan(request)
	loan.IdempotencyKey = idempotencyKey

	var createdLoan model.Loan
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var createErr error
		createdLoan, createErr = s.loanRepository.CreateInTx(tx, loan)
		return createErr
	})
	if errors.Is(err, repository.ErrDuplicateIdempotencyKey) {
		existingLoan, findErr := s.loanRepository.FindByIdempotencyKey(s.db.WithContext(ctx), idempotencyKey)
		if findErr != nil {
			if errors.Is(findErr, gorm.ErrRecordNotFound) {
				return 0, err
			}
			return 0, findErr
		}
		return matchingLoanID(existingLoan, request)
	}
	if err != nil {
		return 0, err
	}

	return createdLoan.ID, nil
}

func matchingLoanID(loan model.Loan, request dto.CreateLoanRequest) (uint64, error) {
	if loan.BorrowerID != request.BorrowerID ||
		!loan.Principal.Equal(request.Principal) ||
		!loan.Rate.Equal(request.Rate) ||
		!loan.ROI.Equal(request.ROI) ||
		loan.AgreementLetterURL != request.AgreementLetterURL {
		return 0, ErrIdempotencyConflict
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
