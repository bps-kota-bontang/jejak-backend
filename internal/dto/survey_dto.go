package dto

import "time"

type CreateSurveyRequest struct {
	SurveyID       string `json:"surveyId" validate:"required"`
	SurveyPeriodID string `json:"surveyPeriodId" validate:"required"`
	XSRFToken      string `json:"xsrfToken" validate:"required"`
	Cookie         string `json:"cookie" validate:"required"`
}

type SyncSurveyAssignmentsRequest struct {
	SurveyPeriodID            string `json:"surveyPeriodId" validate:"required"`
	AssignmentErrorStatusType int    `json:"assignmentErrorStatusType"`
	AssignmentStatusAlias     string `json:"assignmentStatusAlias"`
	FilterTargetType          string `json:"filterTargetType"`
}

type SyncSurveyAssignmentsResponse struct {
	TotalAssignments int `json:"totalAssignments"`
	SavedAssignments int `json:"savedAssignments"`
	SavedLogs        int `json:"savedLogs"`
	SavedAnswers     int `json:"savedAnswers"`
}

type LocationAnswerStat struct {
	CanonicalID string
	Latitude    float64
	Longitude   float64
	AnswerCount int
	Proportion  float64
}

type AssignmentSurveyAnalysis struct {
	AssignmentID   string
	SurveyPeriodID string
	TotalAnswers   int
	Locations      []LocationAnswerStat
}

type SurveyFraudAnalysisResult struct {
	SurveyPeriodID      string
	TotalAssignments    int
	AnalyzedAssignments int
	GeneratedAt         time.Time
	Assignments         []AssignmentSurveyAnalysis
}
