package repositories

import (
	"jejak/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AssignmentRepositoryImpl struct {
	db *gorm.DB
}

func NewAssignmentRepository(db *gorm.DB) AssignmentRepository {
	return &AssignmentRepositoryImpl{db: db}
}

func (r *AssignmentRepositoryImpl) Upsert(assignment *models.Assignment) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "assignment_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"survey_period_id",
			"latitude",
			"longitude",
			"opened_at",
			"submitted_at",
			"revised_at",
			"updated_at",
		}),
	}).Create(assignment).Error
}

func (r *AssignmentRepositoryImpl) FindByAssignmentID(assignmentID string) (*models.Assignment, error) {
	var assignment models.Assignment
	if err := r.db.Where("assignment_id = ?", assignmentID).First(&assignment).Error; err != nil {
		return nil, err
	}
	return &assignment, nil
}

func (r *AssignmentRepositoryImpl) FindBySurveyPeriodID(surveyPeriodID string) ([]models.Assignment, error) {
	var assignments []models.Assignment
	if err := r.db.Where("survey_period_id = ?", surveyPeriodID).Find(&assignments).Error; err != nil {
		return nil, err
	}
	return assignments, nil
}

func (r *AssignmentRepositoryImpl) UpdateViolation(assignmentID string, isViolation bool, violationNote *string, violationScore *float64) error {
	return r.db.Model(&models.Assignment{}).
		Where("assignment_id = ?", assignmentID).
		Updates(map[string]interface{}{
			"is_violation":    isViolation,
			"violation_note":  violationNote,
			"violation_score": violationScore,
		}).Error
}
