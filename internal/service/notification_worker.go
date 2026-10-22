package service

import (
	"context"
	"errors"
	"log"
	"time"

	"loan-service/internal/repository"
)

const (
	notificationWorkerInterval = 10 * time.Minute
	notificationBatchSize      = 200
)

type NotificationWorker struct {
	notificationLogRepository *repository.NotificationLogRepository
	notificationService       *NotificationService
}

func NewNotificationWorker(
	notificationLogRepository *repository.NotificationLogRepository,
	notificationService *NotificationService,
) *NotificationWorker {
	return &NotificationWorker{
		notificationLogRepository: notificationLogRepository,
		notificationService:       notificationService,
	}
}

func (w *NotificationWorker) Run(ctx context.Context) {
	if err := w.selectPendingAndSendEmail(ctx); err != nil {
		log.Printf("notification worker run failed: %v", err)
	}

	ticker := time.NewTicker(notificationWorkerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.selectPendingAndSendEmail(ctx); err != nil {
				log.Printf("notification worker run failed: %v", err)
			}
		}
	}
}

func (w *NotificationWorker) selectPendingAndSendEmail(ctx context.Context) error {
	var processingErrors []error
	for {
		notificationLogs, err := w.notificationLogRepository.ListPending(
			ctx,
			notificationBatchSize,
			MaxRetryCount,
		)
		if err != nil {
			return errors.Join(append(processingErrors, err)...)
		}
		if len(notificationLogs) == 0 {
			return errors.Join(processingErrors...)
		}

		if err := w.notificationService.NotifyLoanInvested(ctx, notificationLogs); err != nil {
			processingErrors = append(processingErrors, err)
			if errors.Is(err, errNotificationStatePersistence) {
				return errors.Join(processingErrors...)
			}
		}
	}
}
