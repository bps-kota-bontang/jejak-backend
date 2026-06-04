package repositories

import (
	"jejak/internal/models"

	"gorm.io/gorm"
)

type AnswerRepositoryImpl struct {
	db *gorm.DB
}

func NewAnswerRepository(db *gorm.DB) AnswerRepository {
	return &AnswerRepositoryImpl{db: db}
}

func (r *AnswerRepositoryImpl) ReplaceByAssignmentID(assignmentID string, answers []models.Answer) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("assignment_id = ?", assignmentID).Delete(&models.Answer{}).Error; err != nil {
			return err
		}
		if len(answers) == 0 {
			return nil
		}
		return tx.Create(&answers).Error
	})
}

func (r *AnswerRepositoryImpl) FindByAssignmentID(assignmentID string) ([]models.Answer, error) {
	var answers []models.Answer
	if err := r.db.Where("assignment_id = ?", assignmentID).Find(&answers).Error; err != nil {
		return nil, err
	}
	return answers, nil
}

func (r *AnswerRepositoryImpl) UpdateLocationID(answerID string, locationID *string) error {
	return r.db.Model(&models.Answer{}).
		Where("id = ?", answerID).
		Update("location_id", locationID).Error
}
