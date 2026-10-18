package repository

import (
	"context"
	"fmt"

	"loan-service/internal/model"

	"gorm.io/gorm"
)

type NotificationLogRepository struct {
	db *gorm.DB
}

func NewNotificationLogRepository(db *gorm.DB) *NotificationLogRepository {
	return &NotificationLogRepository{db: db}
}

func (r *NotificationLogRepository) Create(tx *gorm.DB, notificationLog *model.NotificationLog) error {
	if tx == nil {
		return fmt.Errorf("create notification log: database connection is nil")
	}
	if err := tx.Create(notificationLog).Error; err != nil {
		return fmt.Errorf("create notification log: %w", err)
	}

	return nil
}

func (r *NotificationLogRepository) UpdateStatus(ctx context.Context, id uint64, status model.NotificationStatus) error {
	if r.db == nil {
		return fmt.Errorf("update notification status: database connection is nil")
	}

	result := r.db.WithContext(ctx).
		Model(&model.NotificationLog{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("update notification status: %w", result.Error)
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("update notification status: expected one affected row, got %d", result.RowsAffected)
	}

	return nil
}
