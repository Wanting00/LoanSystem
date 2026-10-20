package service

import (
	"context"
	"errors"
	"strings"

	"loan-service/internal/dto"
	"loan-service/internal/model"
	"loan-service/internal/repository"
	"loan-service/mapper"

	"gorm.io/gorm"
)

var (
	ErrLoanNotFound              = errors.New("loan not found")
	ErrLoanNotProposed           = errors.New("loan is not in proposed status")
	ErrApprovalConflict          = errors.New("approval record already exists")
	ErrInvalidApprovalRequest    = errors.New("approval information is required")
	ErrInvalidValidatorPhotoURL  = errors.New("validator photo URL must link to a PDF or JPEG")
	ErrInvalidAgreementLetterURL = errors.New("agreement letter URL is required and must link to a PDF or JPEG")
	ErrInvalidCreateLoanRequest  = errors.New("borrowerId, principal, rate and roi are required")
	ErrInvalidIdempotencyKey     = errors.New("Idempotency-Key must be 1-255 printable ASCII characters without spaces")
	ErrLoanAlreadyCreated        = errors.New("loan already created")
	ErrIdempotencyConflict       = errors.New("idempotency key was already used with different loan parameters")
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
	if request.BorrowerID == 0 || request.Principal == nil || request.Rate == nil || request.ROI == nil {
		return 0, ErrInvalidCreateLoanRequest
	}
	if !IsValidIdempotencyKey(idempotencyKey) {
		return 0, ErrInvalidIdempotencyKey
	}

	if strings.TrimSpace(request.AgreementLetterURL) == "" || !isPDFOrJPEGURL(request.AgreementLetterURL) {
		return 0, ErrInvalidAgreementLetterURL
	}

	existingLoan, err := s.loanRepository.FindByIdempotencyKey(s.db.WithContext(ctx), idempotencyKey)
	if err == nil {
		return replayedLoan(existingLoan, request)
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
		return replayedLoan(existingLoan, request)
	}
	if err != nil {
		return 0, err
	}

	return createdLoan.ID, nil
}

func IsValidIdempotencyKey(key string) bool {
	if len(key) == 0 || len(key) > 255 {
		return false
	}
	for i := 0; i < len(key); i++ {
		if key[i] < 0x21 || key[i] > 0x7e {
			return false
		}
	}
	return true
}

// replayedLoan reports a reused idempotency key: the original loan ID is still
// returned, with ErrLoanAlreadyCreated when parameters match.
func replayedLoan(loan model.Loan, request dto.CreateLoanRequest) (uint64, error) {
	id, err := matchingLoanID(loan, request)
	if err != nil {
		return 0, err
	}
	return id, ErrLoanAlreadyCreated
}

func matchingLoanID(loan model.Loan, request dto.CreateLoanRequest) (uint64, error) {
	if loan.BorrowerID != request.BorrowerID ||
		!loan.Principal.Equal(*request.Principal) ||
		!loan.Rate.Equal(*request.Rate) ||
		!loan.ROI.Equal(*request.ROI) ||
		loan.AgreementLetterURL != request.AgreementLetterURL {
		return 0, ErrIdempotencyConflict
	}

	return loan.ID, nil
}

func (s *LoanService) GetLoan(ctx context.Context, id uint64) (dto.LoanResponse, error) {
	loan, err := s.loanRepository.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.LoanResponse{}, ErrLoanNotFound
	}
	if err != nil {
		return dto.LoanResponse{}, err
	}

	return mapper.LoanToLoanResponse(loan), nil
}

func (s *LoanService) ApproveLoan(ctx context.Context, id uint64, request dto.ApproveLoanRequest) error {
	if strings.TrimSpace(request.ValidatorPhotoURL) == "" ||
		strings.TrimSpace(request.EmployeeID) == "" ||
		request.ApprovalDate.IsZero() {
		return ErrInvalidApprovalRequest
	}
	if !isPDFOrJPEGURL(request.ValidatorPhotoURL) {
		return ErrInvalidValidatorPhotoURL
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
