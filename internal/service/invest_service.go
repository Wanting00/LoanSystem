package service

import (
	"context"
	"errors"

	"loan-service/internal/dto"
	"loan-service/internal/model"
	"loan-service/internal/repository"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

var (
	ErrInvestorNotFound         = errors.New("investor not found")
	ErrInvalidInvestmentAmount  = errors.New("investment amount must be greater than zero")
	ErrLoanNotApprovedForInvest = errors.New("loan is not in approved status")
	ErrInvestmentExceedsLoan    = errors.New("investment exceeds remaining loan principal")
)

type InvestService struct {
	db                   *gorm.DB
	loanRepository       *repository.LoanRepository
	investmentRepository *repository.InvestmentRepository
	investorRepository   *repository.InvestorRepository
}

func NewInvestService(
	db *gorm.DB,
	loanRepository *repository.LoanRepository,
	investmentRepository *repository.InvestmentRepository,
	investorRepository *repository.InvestorRepository,
) *InvestService {
	return &InvestService{
		db:                   db,
		loanRepository:       loanRepository,
		investmentRepository: investmentRepository,
		investorRepository:   investorRepository,
	}
}

func (s *InvestService) Invest(ctx context.Context, loanID uint64, request dto.InvestRequest) error {
	if request.Amount <= 0 {
		return ErrInvalidInvestmentAmount
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		loan, err := s.loanRepository.FindByIDForUpdate(tx, loanID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrLoanNotFound
		}
		if err != nil {
			return err
		}

		// check if the loan is approved for investment
		if loan.Status != model.LoanStatusApproved {
			return ErrLoanNotApprovedForInvest
		}

		//check if the investor exists
		investorExists, err := s.investorRepository.ExistsByID(tx, request.InvestorID)
		if err != nil {
			return err
		}
		if !investorExists {
			return ErrInvestorNotFound
		}

		// check if the investment exceeds the remaining loan principal
		totalInvested, err := s.investmentRepository.TotalAmountByLoanID(tx, loan.ID)
		if err != nil {
			return err
		}
		amount := decimal.NewFromInt(request.Amount)
		newTotal := totalInvested.Add(amount)
		if newTotal.GreaterThan(loan.Principal) {
			return ErrInvestmentExceedsLoan
		}

		if err := s.investmentRepository.Create(tx, model.Investment{
			LoanID:     loan.ID,
			InvestorID: request.InvestorID,
			Amount:     request.Amount,
		}); err != nil {
			return err
		}

		if newTotal.Equal(loan.Principal) {
			return s.loanRepository.UpdateLoanStatus(tx, loan.ID, model.LoanStatusInvested)
		}

		return nil
	})
}
