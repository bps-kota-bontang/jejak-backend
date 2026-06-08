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
			"region_full_code",
			"region_level1",
			"region_level2",
			"region_level3",
			"region_level4",
			"region_level5",
			"region_level6",
			"latitude",
			"longitude",
			"opened_at",
			"started_at",
			"submitted_at",
			"revised_at",
			"updated_at",
		}),
	}).Create(assignment).Error
}

func (r *AssignmentRepositoryImpl) FindByAssignmentID(assignmentID string) (*models.Assignment, error) {
	var assignment models.Assignment
	if err := r.db.Preload("Locations").
		Where("assignment_id = ?", assignmentID).
		First(&assignment).Error; err != nil {
		return nil, err
	}
	return &assignment, nil
}

func (r *AssignmentRepositoryImpl) FindBySurveyPeriodID(surveyPeriodID string) ([]models.Assignment, error) {
	var assignments []models.Assignment
	if err := r.db.Preload("Locations").
		Where("survey_period_id = ?", surveyPeriodID).
		Find(&assignments).Error; err != nil {
		return nil, err
	}
	return assignments, nil
}

func (r *AssignmentRepositoryImpl) FindBySurveyPeriodIDWithFilter(surveyPeriodID string, filter AssignmentRegionFilter) ([]models.Assignment, error) {
	query := r.db.Where("survey_period_id = ?", surveyPeriodID)

	if filter.RegionFullCode != "" {
		query = query.Where("region_full_code = ?", filter.RegionFullCode)
	}
	if filter.RegionLevel1 != "" {
		query = query.Where("region_level1 = ?", filter.RegionLevel1)
	}
	if filter.RegionLevel2 != "" {
		query = query.Where("region_level2 = ?", filter.RegionLevel2)
	}
	if filter.RegionLevel3 != "" {
		query = query.Where("region_level3 = ?", filter.RegionLevel3)
	}
	if filter.RegionLevel4 != "" {
		query = query.Where("region_level4 = ?", filter.RegionLevel4)
	}
	if filter.RegionLevel5 != "" {
		query = query.Where("region_level5 = ?", filter.RegionLevel5)
	}
	if filter.RegionLevel6 != "" {
		query = query.Where("region_level6 = ?", filter.RegionLevel6)
	}

	var assignments []models.Assignment
	if err := query.Preload("Locations").Find(&assignments).Error; err != nil {
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
