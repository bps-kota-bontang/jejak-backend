package dto

type SystemFeaturesResponse struct {
	FasihAvailable bool `json:"fasih_available"`
}

type SystemFasihAuthorizationResponse struct {
	SurveyPeriodID  string `json:"survey_period_id"`
	FasihAuthorized bool   `json:"fasih_authorized"`
}
