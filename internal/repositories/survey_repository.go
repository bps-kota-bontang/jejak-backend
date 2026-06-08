package repositories

import "jejak/internal/models"

type SurveyRepository interface {
	Create(survey *models.Survey) error
	Upsert(survey *models.Survey) error
	UpdateBySurveyPeriodID(surveyPeriodID string, survey *models.Survey) error
	FindAll() ([]models.Survey, error)
	FindByID(id string) (*models.Survey, error)
	FindBySurveyPeriodID(surveyPeriodID string) (*models.Survey, error)
	UpdateRegionMetadata(surveyPeriodID string, groupID string, levelCount int) error
	UpdateSurveyRegionAssignmentCounts(surveyPeriodID string) error
	ReplaceSurveyRegions(surveyPeriodID string, regions []models.Region) error
	FindSurveyRegionsByLevel(surveyPeriodID string, level int, parentFullCode string) ([]models.Region, error)
	FindBySurveyPeriodIDWithFilter(surveyPeriodID string, filter AssignmentRegionFilter) ([]models.Region, error)
}
