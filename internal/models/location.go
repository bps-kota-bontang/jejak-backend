package models

import (
	"time"

	"gorm.io/gorm"
)

type Location struct {
	ID                     string     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AssignmentID           string     `gorm:"type:text;not null;index:idx_assignment_canonical_id,priority:1"`
	Assignment             Assignment `gorm:"foreignKey:AssignmentID;references:AssignmentID"`
	CanonicalID            string     `gorm:"type:text;not null;index:idx_assignment_canonical_id,priority:2"`
	Latitude               float64    `gorm:"type:double precision;not null"`
	Longitude              float64    `gorm:"type:double precision;not null"`
	AnswerCount            int        `gorm:"not null;default:0"`
	Proportion             float64    `gorm:"type:double precision;not null;default:0"`
	DistanceToSampleMeters float64    `gorm:"type:double precision;not null;default:0"`
	WithinSampleAreaRadius bool       `gorm:"not null;default:false"`
	CreatedAt              time.Time
	UpdatedAt              time.Time
	DeletedAt              gorm.DeletedAt `gorm:"index"`
}
