package repositories

import "jejak/internal/models"

type LocationRepository interface {
	ReplaceByAssignmentID(assignmentID string, locations []models.Location) error
	FindByAssignmentID(assignmentID string) ([]models.Location, error)
}
