package repositories

import (
	"jejak/internal/models"

	"gorm.io/gorm"
)

type LogRepositoryImpl struct {
	db *gorm.DB
}

func NewLogRepository(db *gorm.DB) LogRepository {
	return &LogRepositoryImpl{db: db}
}

func (r *LogRepositoryImpl) ReplaceByAssignmentID(assignmentID string, logs []models.Log) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("assignment_id = ?", assignmentID).Delete(&models.Log{}).Error; err != nil {
			return err
		}
		if len(logs) == 0 {
			return nil
		}
		return tx.Create(&logs).Error
	})
}

func (r *LogRepositoryImpl) FindByAssignmentID(assignmentID string) ([]models.Log, error) {
	var logs []models.Log
	if err := r.db.Where("assignment_id = ?", assignmentID).Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
