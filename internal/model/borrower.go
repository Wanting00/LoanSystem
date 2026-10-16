package model

import "time"

type Borrower struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement"`
	IDNumber  string `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Loans []Loan `gorm:"foreignKey:BorrowerID"`
}
