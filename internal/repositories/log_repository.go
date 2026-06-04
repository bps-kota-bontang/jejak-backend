package repositories

import "jejak/internal/models"

type LogRepository interface {
	ReplaceByAssignmentID(assignmentID string, logs []models.Log) error
	FindByAssignmentID(assignmentID string) ([]models.Log, error)
}
