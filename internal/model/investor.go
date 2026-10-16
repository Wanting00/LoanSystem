package model

import "time"

type Investor struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement"`
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time

	Investments []Investment `gorm:"foreignKey:InvestorID"`
}
