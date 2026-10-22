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

func (r *NotificationLogRepository) ListPending(
	ctx context.Context,
	limit int,
	maxRetryCount int,
) ([]model.NotificationLog, error) {
	if r.db == nil {
		return nil, fmt.Errorf("list pending notification logs: database connection is nil")
	}

	query := r.db.WithContext(ctx).
		Where("status = ? AND retry_count <= ?", model.NotificationStatusPending, maxRetryCount)

	var notificationLogs []model.NotificationLog
	if err := query.
		Order("created_at ASC").
		Order("id ASC").
		Limit(limit).
		Find(&notificationLogs).Error; err != nil {
		return nil, fmt.Errorf("list pending notification logs: %w", err)
	}

	return notificationLogs, nil
}

func (r *NotificationLogRepository) TryMarkSending(ctx context.Context, id uint64, maxRetryCount int) (bool, error) {
	if r.db == nil {
		return false, fmt.Errorf("claim notification log: database connection is nil")
	}

	result := r.db.WithContext(ctx).
		Model(&model.NotificationLog{}).
		Where("id = ? AND status = ? AND retry_count <= ?", id, model.NotificationStatusPending, maxRetryCount).
		Update("status", model.NotificationStatusSending)
	if result.Error != nil {
		return false, fmt.Errorf("claim notification log %d: %w", id, result.Error)
	}

	return result.RowsAffected == 1, nil
}

func (r *NotificationLogRepository) MarkSent(ctx context.Context, id uint64) error {
	if r.db == nil {
		return fmt.Errorf("mark notification sent: database connection is nil")
	}

	result := r.db.WithContext(ctx).
		Model(&model.NotificationLog{}).
		Where("id = ? AND status = ?", id, model.NotificationStatusSending).
		Update("status", model.NotificationStatusSent)
	if result.Error != nil {
		return fmt.Errorf("mark notification log %d sent: %w", id, result.Error)
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("mark notification log %d sent: expected one affected row, got %d", id, result.RowsAffected)
	}

	return nil
}

func (r *NotificationLogRepository) RecordFailure(
	ctx context.Context,
	id uint64,
	retryCount int,
	status model.NotificationStatus,
) error {
	if r.db == nil {
		return fmt.Errorf("record notification failure: database connection is nil")
	}

	result := r.db.WithContext(ctx).
		Model(&model.NotificationLog{}).
		Where("id = ? AND status = ?", id, model.NotificationStatusSending).
		Updates(map[string]interface{}{
			"retry_count": retryCount,
			"status":      status,
		})
	if result.Error != nil {
		return fmt.Errorf("record failure for notification log %d: %w", id, result.Error)
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("record failure for notification log %d: expected one affected row, got %d", id, result.RowsAffected)
	}

	return nil
}
