package model

import "time"

type Borrower struct {
	ID                 uint64 `gorm:"primaryKey;autoIncrement"`
	IdentityCardNumber string `gorm:"type:varchar(64);uniqueIndex;not null"`
	CreatedAt          time.Time
	UpdatedAt          time.Time

	Loans []Loan `gorm:"foreignKey:BorrowerID"`
}
