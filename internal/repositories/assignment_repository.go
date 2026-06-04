package repositories

import "jejak/internal/models"

type AssignmentRepository interface {
	Upsert(assignment *models.Assignment) error
	FindByAssignmentID(assignmentID string) (*models.Assignment, error)
	FindBySurveyPeriodID(surveyPeriodID string) ([]models.Assignment, error)
	UpdateViolation(assignmentID string, isViolation bool, violationNote *string, violationScore *float64) error
}
