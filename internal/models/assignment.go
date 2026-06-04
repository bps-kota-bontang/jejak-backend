package models

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type Assignment struct {
	ID             string  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SurveyPeriodID string  `gorm:"type:text;not null"`        // c7ee8024-e2fa-4d4a-b463-85f7ee508702
	AssignmentID   string  `gorm:"type:text;unique;not null"` // 1c0e2e43-fb41-4d1f-bb52-3e6cded8c343
	Latitude       float64 `gorm:"type:double precision;not null"`
	Longitude      float64 `gorm:"type:double precision;not null"`
	OpenedAt       sql.NullTime
	SubmittedAt    time.Time
	RevisedAt      time.Time
	IsViolation    bool    `gorm:"not null;default:false"`
	ViolationNote  *string `gorm:"type:text"`
	ViolationScore *float64 `gorm:"type:double precision"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}
