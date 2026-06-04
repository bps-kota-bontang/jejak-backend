package repositories

import "jejak/internal/models"

type AnswerRepository interface {
	ReplaceByAssignmentID(assignmentID string, answers []models.Answer) error
	FindByAssignmentID(assignmentID string) ([]models.Answer, error)
	UpdateLocationID(answerID string, locationID *string) error
}
