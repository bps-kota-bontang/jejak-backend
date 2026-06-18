package models

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Log struct {
	ID           string  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AssignmentID string  `gorm:"type:text;not null;uniqueIndex:idx_logs_assignment_event_hash,priority:1"` // 1c0e2e43-fb41-4d1f-bb52-3e6cded8c343
	EventHash    string  `gorm:"type:char(64);uniqueIndex:idx_logs_assignment_event_hash,priority:2"`
	Action       string  `gorm:"type:text;not null"` // opened, submitted, revised
	Latitude     float64 `gorm:"type:double precision;not null"`
	Longitude    float64 `gorm:"type:double precision;not null"`
	ActionedAt   time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func BuildLogEventHash(assignmentID, action string, latitude, longitude float64, actionedAt time.Time) string {
	normalized := strings.Join([]string{
		strings.TrimSpace(assignmentID),
		strings.TrimSpace(action),
		strconv.FormatFloat(latitude, 'g', -1, 64),
		strconv.FormatFloat(longitude, 'g', -1, 64),
		actionedAt.UTC().Format(time.RFC3339Nano),
	}, "|")

	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}
