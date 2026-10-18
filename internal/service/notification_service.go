package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"loan-service/internal/model"
	"loan-service/internal/repository"

	"gorm.io/gorm"
)

func sendEmail(recipient, agreementLetterURL string) error {
	log.Printf("mock email sent to %s with agreement letter %s", recipient, agreementLetterURL)
	return nil
}

type NotificationService struct {
	db                        *gorm.DB
	investmentRepository      *repository.InvestmentRepository
	investorRepository        *repository.InvestorRepository
	notificationLogRepository *repository.NotificationLogRepository
}

func NewNotificationService(
	db *gorm.DB,
	investmentRepository *repository.InvestmentRepository,
	investorRepository *repository.InvestorRepository,
	notificationLogRepository *repository.NotificationLogRepository,
) *NotificationService {
	return &NotificationService{
		db:                        db,
		investmentRepository:      investmentRepository,
		investorRepository:        investorRepository,
		notificationLogRepository: notificationLogRepository,
	}
}

func (s *NotificationService) NotifyLoanInvested(ctx context.Context, loanID uint64, agreementLetterURL string) error {
	investments, err := s.investmentRepository.ListByLoanID(s.db.WithContext(ctx), loanID)
	if err != nil {
		return fmt.Errorf("list investments for loan %d: %w", loanID, err)
	}

	var notificationErrors []error
	for _, investment := range investments {
		recipient, err := s.investorRepository.GetEmailByID(s.db.WithContext(ctx), investment.InvestorID)
		if err != nil {
			notificationErrors = append(notificationErrors, fmt.Errorf("get email for investor %d: %w", investment.InvestorID, err))
			continue
		}

		notificationLog := &model.NotificationLog{
			LoanID:             loanID,
			InvestmentID:       investment.ID,
			Channel:            model.ChannelEmail,
			Recipient:          recipient,
			AgreementLetterURL: agreementLetterURL,
			Status:             model.NotificationStatusPending,
		}
		if err := s.notificationLogRepository.Create(s.db.WithContext(ctx), notificationLog); err != nil {
			notificationErrors = append(notificationErrors, fmt.Errorf("create notification log for investment %d: %w", investment.ID, err))
			continue
		}

		sendErr := sendEmail(recipient, agreementLetterURL)
		status := model.NotificationStatusSent
		if sendErr != nil {
			status = model.NotificationStatusFailed
			notificationErrors = append(notificationErrors, fmt.Errorf("send notification for investment %d: %w", investment.ID, sendErr))
		}
		if err := s.notificationLogRepository.UpdateStatus(ctx, notificationLog.ID, status); err != nil {
			notificationErrors = append(notificationErrors, fmt.Errorf("update notification log %d: %w", notificationLog.ID, err))
		}
	}

	return errors.Join(notificationErrors...)
}
