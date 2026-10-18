package model

import "time"

type Investor struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement"`
	Name      string
	Email     string `gorm:"type:varchar(255);not null;default:''"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Investments []Investment `gorm:"foreignKey:InvestorID"`
}
