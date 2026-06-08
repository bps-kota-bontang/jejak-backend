package repositories

import (
	"jejak/internal/models"
	"time"

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

func (r *LogRepositoryImpl) FindBySurveyPeriodIDRegionFullCodeAndActionedAt(
	surveyPeriodID string,
	regionFullCode string,
	actionedAtFrom *time.Time,
	actionedAtTo *time.Time,
) ([]models.Log, error) {
	query := r.db.Model(&models.Log{}).
		Joins("JOIN assignments ON assignments.assignment_id = logs.assignment_id").
		Where("assignments.survey_period_id = ?", surveyPeriodID).
		Where("assignments.region_full_code = ?", regionFullCode)

	if actionedAtFrom != nil {
		query = query.Where("logs.actioned_at >= ?", *actionedAtFrom)
	}
	if actionedAtTo != nil {
		query = query.Where("logs.actioned_at <= ?", *actionedAtTo)
	}

	var logs []models.Log
	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}
