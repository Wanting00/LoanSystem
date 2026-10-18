package service

import "context"

type NotificationService struct{}

func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

func (s *NotificationService) CreateTasks(ctx context.Context, loanID uint64) error {
	return nil
}
