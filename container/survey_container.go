package container

import "jejak/internal/services"

type SurveyContainer struct {
	SurveyService *services.SurveyService
}

func NewSurveyContainer(surveyService *services.SurveyService) *SurveyContainer {
	return &SurveyContainer{SurveyService: surveyService}
}
