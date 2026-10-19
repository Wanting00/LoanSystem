package service

import (
	"context"
	"errors"
	"strings"

	"loan-service/internal/dto"
	"loan-service/internal/model"
	"loan-service/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrInvalidDisbursementRequest = errors.New("signed agreement URL, employee ID, and valid disbursement date are required")
	ErrInvalidSignedAgreementURL  = errors.New("signed agreement URL must link to a PDF or JPEG")
	ErrLoanNotInvested            = errors.New("loan is not in invested status")
	ErrDisbursementConflict       = errors.New("loan has already been disbursed")
)

type DisburseService struct {
	db                     *gorm.DB
	loanRepository         *repository.LoanRepository
	disbursementRepository *repository.DisbursementRepository
}

func NewDisburseService(
	db *gorm.DB,
	loanRepository *repository.LoanRepository,
	disbursementRepository *repository.DisbursementRepository,
) *DisburseService {
	return &DisburseService{
		db:                     db,
		loanRepository:         loanRepository,
		disbursementRepository: disbursementRepository,
	}
}

func (s *DisburseService) Disburse(
	ctx context.Context,
	loanID uint64,
	request dto.DisburseRequest,
) (uint64, error) {
	if loanID == 0 {
		return 0, ErrInvalidDisbursementRequest
	}

	signedAgreementURL := strings.TrimSpace(request.SignedAgreementURL)
	employeeID := strings.TrimSpace(request.EmployeeID)
	if signedAgreementURL == "" || employeeID == "" || request.DisbursementDate.IsZero() {
		return 0, ErrInvalidDisbursementRequest
	}
	if !isPDFOrJPEGURL(signedAgreementURL) {
		return 0, ErrInvalidSignedAgreementURL
	}

	var disbursement model.Disbursement
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		loan, err := s.loanRepository.FindByIDForUpdate(tx, loanID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrLoanNotFound
		}
		if err != nil {
			return err
		}
		if loan.Status == model.LoanStatusDisbursed {
			return ErrDisbursementConflict
		}
		if loan.Status != model.LoanStatusInvested {
			return ErrLoanNotInvested
		}

		disbursement = model.Disbursement{
			LoanID:             loan.ID,
			SignedAgreementURL: signedAgreementURL,
			EmployeeID:         employeeID,
			DisbursementDate:   request.DisbursementDate,
		}
		if err := s.disbursementRepository.Create(tx, &disbursement); err != nil {
			if errors.Is(err, repository.ErrDuplicateDisbursement) {
				return ErrDisbursementConflict
			}
			return err
		}
		if err := s.loanRepository.UpdateLoanStatus(tx, loan.ID, model.LoanStatusDisbursed); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return disbursement.ID, nil
}
