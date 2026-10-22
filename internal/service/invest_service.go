package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"loan-service/internal/dto"
	"loan-service/internal/model"
	"loan-service/internal/repository"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

var (
	ErrInvestorNotFound         = errors.New("investor not found")
	ErrInvalidInvestRequest     = errors.New("investorId and amount are required")
	ErrInvalidInvestmentAmount  = errors.New("investment amount must be greater than zero")
	ErrLoanNotApprovedForInvest = errors.New("loan is not in approved status")
	ErrInvestmentExceedsLoan    = errors.New("investment exceeds remaining loan principal")
)

type InvestService struct {
	db                   *gorm.DB
	loanRepository       *repository.LoanRepository
	investmentRepository *repository.InvestmentRepository
	investorRepository   *repository.InvestorRepository
	notificationLogRepo  *repository.NotificationLogRepository
	notificationService  *NotificationService
}

func NewInvestService(
	db *gorm.DB,
	loanRepository *repository.LoanRepository,
	investmentRepository *repository.InvestmentRepository,
	investorRepository *repository.InvestorRepository,
	notificationLogRepo *repository.NotificationLogRepository,
	notificationService *NotificationService,
) *InvestService {
	return &InvestService{
		db:                   db,
		loanRepository:       loanRepository,
		investmentRepository: investmentRepository,
		investorRepository:   investorRepository,
		notificationLogRepo:  notificationLogRepo,
		notificationService:  notificationService,
	}
}

func (s *InvestService) Invest(ctx context.Context, loanID uint64, request dto.InvestRequest) error {
	if request.InvestorID == 0 || request.Amount == nil {
		return ErrInvalidInvestRequest
	}
	amount := *request.Amount
	if !amount.GreaterThan(decimal.Zero) || !amount.Equal(amount.Truncate(2)) {
		return ErrInvalidInvestmentAmount
	}

	var loanBecameInvested bool
	var notificationLogs []model.NotificationLog

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
		newTotal := totalInvested.Add(amount)
		if newTotal.GreaterThan(loan.Principal) {
			return ErrInvestmentExceedsLoan
		}

		investment := &model.Investment{
			LoanID:     loan.ID,
			InvestorID: request.InvestorID,
			Amount:     amount,
		}
		if err := s.investmentRepository.Create(tx, investment); err != nil {
			return err
		}

		if newTotal.Equal(loan.Principal) {
			if err := s.loanRepository.UpdateLoanStatus(tx, loan.ID, model.LoanStatusInvested); err != nil {
				return err
			}
			loanBecameInvested = true

			investments, err := s.investmentRepository.ListByLoanID(tx, loan.ID)
			if err != nil {
				return err
			}

			// create notification logs(for email) for each investment
			for _, investment := range investments {
				recipient, err := s.investorRepository.GetEmailByID(tx, investment.InvestorID)
				if err != nil {
					return err
				}

				notificationLog := model.NotificationLog{
					EventKey:           fmt.Sprintf("loan:%d:investment:%d:email", loan.ID, investment.ID),
					LoanID:             loan.ID,
					InvestmentID:       investment.ID,
					Channel:            model.ChannelEmail,
					Recipient:          recipient,
					AgreementLetterURL: loan.AgreementLetterURL,
					Status:             model.NotificationStatusPending,
				}
				if err := s.notificationLogRepo.Create(tx, &notificationLog); err != nil {
					return err
				}
				notificationLogs = append(notificationLogs, notificationLog)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	// notify investors if the loan became fully invested
	if loanBecameInvested {
		go func(logs []model.NotificationLog) {
			notificationCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := s.notificationService.NotifyLoanInvested(notificationCtx, logs); err != nil {
				log.Printf("notify investors for loan %d failed: %v", loanID, err)
			}
		}(notificationLogs)
	}

	return nil
}
