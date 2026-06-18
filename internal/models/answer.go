package models

import (
	"time"

	"gorm.io/gorm"
)

type Answer struct {
	ID           string     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AssignmentID string     `gorm:"type:text;not null"` // 1c0e2e43-fb41-4d1f-bb52-3e6cded8c343
	Assignment   Assignment `gorm:"foreignKey:AssignmentID;references:AssignmentID"`
	LocationID   *string    `gorm:"type:uuid;index"`
	Location     *Location  `gorm:"foreignKey:LocationID;references:ID"`
	Name         string     `gorm:"type:text;not null"` // data1, data2, ..., data10
	AnsweredAt   time.Time
	RevisedAt    time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}
