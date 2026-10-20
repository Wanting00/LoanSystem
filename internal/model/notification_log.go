package model

import "time"

type NotificationChannel string

const (
	ChannelEmail NotificationChannel = "email"
)

type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "pending"
	NotificationStatusSending NotificationStatus = "sending"
	NotificationStatusSent    NotificationStatus = "sent"
	NotificationStatusFailed  NotificationStatus = "failed"
)

type NotificationLog struct {
	ID                 uint64              `gorm:"primaryKey;autoIncrement"`
	LoanID             uint64              `gorm:"not null"`
	InvestmentID       uint64              `gorm:"not null"`
	Channel            NotificationChannel `gorm:"type:varchar(16);not null"`
	Recipient          string              `gorm:"type:varchar(255);not null"`
	AgreementLetterURL string              `gorm:"type:varchar(2048)"`
	Status             NotificationStatus  `gorm:"type:varchar(16);not null;default:pending;index:idx_status"`
	CreatedAt          time.Time           `gorm:"type:datetime(3);not null"`
	UpdatedAt          time.Time           `gorm:"type:datetime(3);not null"`
}
