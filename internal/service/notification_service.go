package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"loan-service/internal/model"
	"loan-service/internal/repository"
)

const MaxRetryCount = 3

var errNotificationStatePersistence = errors.New("notification state persistence failed")

func sendEmail(recipient, agreementLetterURL string) error {
	log.Printf("mock email sent to %s with agreement letter %s", recipient, agreementLetterURL)
	return nil
}

type NotificationService struct {
	notificationLogRepository *repository.NotificationLogRepository
}

func NewNotificationService(
	notificationLogRepository *repository.NotificationLogRepository,
) *NotificationService {
	return &NotificationService{
		notificationLogRepository: notificationLogRepository,
	}
}

func (s *NotificationService) NotifyLoanInvested(ctx context.Context, notificationLogs []model.NotificationLog) error {
	var notificationErrors []error
	for _, notificationLog := range notificationLogs {
		// status: pending → sending (to avoid workers sending the same notification)
		claimed, err := s.notificationLogRepository.TryMarkSending(ctx, notificationLog.ID, MaxRetryCount)
		if err != nil {
			notificationErrors = append(notificationErrors, fmt.Errorf(
				"%w: claim notification %d: %w",
				errNotificationStatePersistence,
				notificationLog.ID,
				err,
			))
			continue
		}
		if !claimed {
			continue
		}

		sendErr := sendEmail(notificationLog.Recipient, notificationLog.AgreementLetterURL)
		if sendErr == nil {
			if err := s.notificationLogRepository.MarkSent(ctx, notificationLog.ID); err != nil {
				notificationErrors = append(notificationErrors, fmt.Errorf(
					"%w: mark notification %d sent: %w",
					errNotificationStatePersistence,
					notificationLog.ID,
					err,
				))
			}
			continue
		}

		// update the retry count and status based on the result of sending the notification
		nextRetryCount := notificationLog.RetryCount + 1
		nextStatus := model.NotificationStatusPending
		if nextRetryCount > MaxRetryCount {
			nextStatus = model.NotificationStatusFailed
		}
		if err := s.notificationLogRepository.RecordFailure(ctx, notificationLog.ID, nextRetryCount, nextStatus); err != nil {
			notificationErrors = append(notificationErrors, errors.Join(
				fmt.Errorf("send notification %d: %w", notificationLog.ID, sendErr),
				fmt.Errorf("%w: record failure for notification %d: %w", errNotificationStatePersistence, notificationLog.ID, err),
			))
			continue
		}
		notificationErrors = append(notificationErrors, fmt.Errorf("send notification %d: %w", notificationLog.ID, sendErr))
	}

	return errors.Join(notificationErrors...)
}
