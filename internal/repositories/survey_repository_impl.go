package repositories

import (
	"jejak/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SurveyRepositoryImpl struct {
	db *gorm.DB
}

func NewSurveyRepository(db *gorm.DB) SurveyRepository {
	return &SurveyRepositoryImpl{db: db}
}

func (r *SurveyRepositoryImpl) Create(survey *models.Survey) error {
	return r.db.Create(survey).Error
}

func (r *SurveyRepositoryImpl) Upsert(survey *models.Survey) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "survey_period_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"survey_id",
			"xsrf_token",
			"cookie",
			"updated_at",
		}),
	}).Create(survey).Error
}

func (r *SurveyRepositoryImpl) FindByID(id string) (*models.Survey, error) {
	var survey models.Survey
	if err := r.db.Where("id = ?", id).First(&survey).Error; err != nil {
		return nil, err
	}
	return &survey, nil
}

func (r *SurveyRepositoryImpl) FindBySurveyPeriodID(surveyPeriodID string) (*models.Survey, error) {
	var survey models.Survey
	if err := r.db.Where("survey_period_id = ?", surveyPeriodID).First(&survey).Error; err != nil {
		return nil, err
	}
	return &survey, nil
}
