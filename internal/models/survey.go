package models

import (
	"time"

	"gorm.io/gorm"
)

//https://fasih-sm.bps.go.id/app/surveys/2561fe7e-c7b6-4e4a-b2d9-c8d254cba2bd/c7ee8024-e2fa-4d4a-b463-85f7ee508702

type Survey struct {
	ID             string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SurveyID       string `gorm:"type:text;unique;not null"` //2561fe7e-c7b6-4e4a-b2d9-c8d254cba2bd
	SurveyPeriodID string `gorm:"type:text;unique;not null"` //c7ee8024-e2fa-4d4a-b463-85f7ee508702
	XSRFToken      string `gorm:"type:text;not null"`
	Cookie         string `gorm:"type:text;not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}
