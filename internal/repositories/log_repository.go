package repositories

import (
	"jejak/internal/models"
	"time"
)

type LogRepository interface {
	ReplaceByAssignmentID(assignmentID string, logs []models.Log) error
	FindByAssignmentID(assignmentID string) ([]models.Log, error)
	FindBySurveyPeriodIDRegionFullCodeAndActionedAt(
		surveyPeriodID string,
		regionFullCode string,
		actionedAtFrom *time.Time,
		actionedAtTo *time.Time,
	) ([]models.Log, error)
}
