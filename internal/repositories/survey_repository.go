package repositories

import "jejak/internal/models"

type RegionLevelOption struct {
	Value string
	Label string
}

type RegionLevelFilter struct {
	Level1 string
	Level2 string
	Level3 string
	Level4 string
	Level5 string
}

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
	FindBySurveyPeriodIDWithFilterPaginated(surveyPeriodID string, filter AssignmentRegionFilter, limit int, offset int) ([]models.Region, error)
	CountBySurveyPeriodIDWithFilter(surveyPeriodID string, filter AssignmentRegionFilter) (int64, error)
	GetDistinctRegionLevel1(surveyPeriodID string) ([]RegionLevelOption, error)
	GetDistinctRegionLevel2(surveyPeriodID string, filter RegionLevelFilter) ([]RegionLevelOption, error)
	GetDistinctRegionLevel3(surveyPeriodID string, filter RegionLevelFilter) ([]RegionLevelOption, error)
	GetDistinctRegionLevel4(surveyPeriodID string, filter RegionLevelFilter) ([]RegionLevelOption, error)
	GetDistinctRegionLevel5(surveyPeriodID string, filter RegionLevelFilter) ([]RegionLevelOption, error)
	GetDistinctRegionLevel6(surveyPeriodID string, filter RegionLevelFilter) ([]RegionLevelOption, error)
}
