package repositories

import "jejak/internal/models"

type SurveyRepository interface {
	Create(survey *models.Survey) error
	Upsert(survey *models.Survey) error
	FindByID(id string) (*models.Survey, error)
	FindBySurveyPeriodID(surveyPeriodID string) (*models.Survey, error)
}
