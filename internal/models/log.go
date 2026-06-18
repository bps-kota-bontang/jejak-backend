package models

import (
	"time"

	"gorm.io/gorm"
)

type Log struct {
	ID           string     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AssignmentID string     `gorm:"type:text;not null"` // 1c0e2e43-fb41-4d1f-bb52-3e6cded8c343
	Assignment   Assignment `gorm:"foreignKey:AssignmentID;references:AssignmentID"`
	Action       string     `gorm:"type:text;not null"` // opened, submitted, revised
	Latitude     float64    `gorm:"type:double precision;not null"`
	Longitude    float64    `gorm:"type:double precision;not null"`
	ActionedAt   time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}
