package repositories

import "jejak/internal/models"

type AssignmentRegionFilter struct {
	RegionFullCode string
	RegionLevel1   string
	RegionLevel2   string
	RegionLevel3   string
	RegionLevel4   string
	RegionLevel5   string
	RegionLevel6   string
	PJ             string
	PML            string
	PPL            string
	Assignment     string
	Status         string
	SortBy         string
	SortDir        string
}

type AssignmentRepository interface {
	Upsert(assignment *models.Assignment) error
	FindByAssignmentID(assignmentID string) (*models.Assignment, error)
	FindBySurveyPeriodID(surveyPeriodID string) ([]models.Assignment, error)
	FindBySurveyPeriodIDWithFilter(surveyPeriodID string, filter AssignmentRegionFilter) ([]models.Assignment, error)
	UpdateViolation(assignmentID string, isViolation bool, violationNote *string, violationScore *float64) error
}
