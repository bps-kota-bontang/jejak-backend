package models

import (
	"database/sql"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	AssignmentStatusDraft     = 1
	AssignmentStatusSubmitted = 2
	AssignmentStatusApproved  = 3
	AssignmentStatusRejected  = 4
	AssignmentStatusRevoked   = 5
)

func AssignmentStatusCodeFromAlias(alias string) *int {
	switch strings.ToUpper(strings.TrimSpace(alias)) {
	case "DRAFT":
		return assignmentStatusPtr(AssignmentStatusDraft)
	case "SUBMITTED BY PENCACAH":
		return assignmentStatusPtr(AssignmentStatusSubmitted)
	case "APPROVED BY PENGAWAS":
		return assignmentStatusPtr(AssignmentStatusApproved)
	case "REJECTED BY PENGAWAS":
		return assignmentStatusPtr(AssignmentStatusRejected)
	case "REVOKED BY PENGAWAS":
		return assignmentStatusPtr(AssignmentStatusRevoked)
	default:
		return nil
	}
}

func AssignmentStatusCodeFromInt(code *int) *int {
	if code == nil {
		return nil
	}

	switch *code {
	case AssignmentStatusDraft, AssignmentStatusSubmitted, AssignmentStatusApproved, AssignmentStatusRejected, AssignmentStatusRevoked:
		return assignmentStatusPtr(*code)
	default:
		return nil
	}
}

func assignmentStatusPtr(code int) *int {
	return &code
}

type Assignment struct {
	ID             string  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SurveyPeriodID string  `gorm:"type:text;not null"`        // c7ee8024-e2fa-4d4a-b463-85f7ee508702
	AssignmentID   string  `gorm:"type:text;unique;not null"` // 1c0e2e43-fb41-4d1f-bb52-3e6cded8c343
	Status         *int    `gorm:"type:int"`
	RegionFullCode *string `gorm:"type:text"`
	RegionLevel1   *string `gorm:"type:text"`
	RegionLevel2   *string `gorm:"type:text"`
	RegionLevel3   *string `gorm:"type:text"`
	RegionLevel4   *string `gorm:"type:text"`
	RegionLevel5   *string `gorm:"type:text"`
	RegionLevel6   *string `gorm:"type:text"`
	Latitude       float64 `gorm:"type:double precision;not null"`
	Longitude      float64 `gorm:"type:double precision;not null"`
	Usaha          int     `gorm:"type:int;not null;default:0"`
	OpenedAt       sql.NullTime
	StartedAt      sql.NullTime
	SubmittedAt    time.Time
	RevisedAt      time.Time
	IsViolation    bool       `gorm:"not null;default:false"`
	ViolationNote  *string    `gorm:"type:text"`
	ViolationScore *float64   `gorm:"type:double precision"`
	Answers        []Answer   `gorm:"foreignKey:AssignmentID;references:AssignmentID"`
	Locations      []Location `gorm:"foreignKey:AssignmentID;references:AssignmentID"`
	Logs           []Log      `gorm:"foreignKey:AssignmentID;references:AssignmentID"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}
