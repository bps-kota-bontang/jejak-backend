package mappers

import (
	"database/sql"
	"jejak/internal/dto"
	"jejak/internal/models"
	"time"
)

func ToSurveyResponse(survey *models.Survey) *dto.SurveyResponse {
	areaResp := &dto.AreaResponse{
		ID:              survey.Area.ID,
		Name:            survey.Area.Name,
		GeoJSONFilePath: survey.Area.GeoJSONFilePath,
		Description:     survey.Area.Description,
		CreatedAt:       survey.Area.CreatedAt,
		UpdatedAt:       survey.Area.UpdatedAt,
	}

	return &dto.SurveyResponse{
		ID:               survey.ID,
		Name:             survey.Name,
		SurveyID:         survey.SurveyID,
		SurveyPeriodID:   survey.SurveyPeriodID,
		XSRFToken:        survey.XSRFToken,
		Cookie:           survey.Cookie,
		RegionLevel1:     survey.RegionLevel1,
		RegionLevel2:     survey.RegionLevel2,
		LogDeltaMaxMins:  survey.LogDeltaMaxMins,
		LogDateFrom:      formatDatePtr(survey.LogDateFrom),
		LogDateTo:        formatDatePtr(survey.LogDateTo),
		RegionGroupID:    survey.RegionGroupID,
		RegionLevelCount: survey.RegionLevelCount,
		AreaID:           survey.AreaID,
		GeoJSONKey:       survey.GeoJSONKey,
		Area:             areaResp,
		CreatedAt:        survey.CreatedAt,
		UpdatedAt:        survey.UpdatedAt,
	}
}

func formatDatePtr(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := value.Format("2006-01-02")
	return &formatted
}

func ToSurveyResponses(surveys []models.Survey) []dto.SurveyResponse {
	responses := make([]dto.SurveyResponse, 0, len(surveys))
	for _, survey := range surveys {
		responses = append(responses, *ToSurveyResponse(&survey))
	}
	return responses
}

func ToAssignmentResponse(assignment *models.Assignment) *dto.AssignmentResponse {
	locations := make([]dto.LocationAnswerStat, 0, len(assignment.Locations))
	for _, location := range assignment.Locations {
		locations = append(locations, dto.LocationAnswerStat{
			CanonicalID:            location.CanonicalID,
			Latitude:               location.Latitude,
			Longitude:              location.Longitude,
			AnswerCount:            location.AnswerCount,
			Proportion:             location.Proportion,
			DistanceToSampleMeters: location.DistanceToSampleMeters,
			WithinSampleAreaRadius: location.WithinSampleAreaRadius,
		})
	}

	return &dto.AssignmentResponse{
		ID:             assignment.ID,
		SurveyPeriodID: assignment.SurveyPeriodID,
		AssignmentID:   assignment.AssignmentID,
		Status:         assignment.Status,
		RegionFullCode: assignment.RegionFullCode,
		RegionLevel1:   assignment.RegionLevel1,
		RegionLevel2:   assignment.RegionLevel2,
		RegionLevel3:   assignment.RegionLevel3,
		RegionLevel4:   assignment.RegionLevel4,
		RegionLevel5:   assignment.RegionLevel5,
		RegionLevel6:   assignment.RegionLevel6,
		Latitude:       assignment.Latitude,
		Longitude:      assignment.Longitude,
		Usaha:          assignment.Usaha,
		OpenedAt:       toTimePtr(assignment.OpenedAt),
		StartedAt:      toTimePtr(assignment.StartedAt),
		SubmittedAt:    assignment.SubmittedAt,
		RevisedAt:      assignment.RevisedAt,
		IsViolation:    assignment.IsViolation,
		ViolationNote:  assignment.ViolationNote,
		ViolationScore: assignment.ViolationScore,
		Locations:      locations,
		CreatedAt:      assignment.CreatedAt,
		UpdatedAt:      assignment.UpdatedAt,
	}
}

func ToAssignmentResponses(assignments []models.Assignment) []dto.AssignmentResponse {
	responses := make([]dto.AssignmentResponse, 0, len(assignments))
	for _, assignment := range assignments {
		responses = append(responses, *ToAssignmentResponse(&assignment))
	}
	return responses
}

func ToAssignmentLogPointResponses(logs []models.Log) []dto.AssignmentLogPointResponse {
	responses := make([]dto.AssignmentLogPointResponse, 0, len(logs))
	for _, item := range logs {
		responses = append(responses, dto.AssignmentLogPointResponse{
			ID:           item.ID,
			AssignmentID: item.AssignmentID,
			Action:       item.Action,
			Latitude:     item.Latitude,
			Longitude:    item.Longitude,
			ActionedAt:   item.ActionedAt,
		})
	}
	return responses
}

func toTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}

func ToSurveyRegionResponse(region *models.Region) *dto.SurveyRegionResponse {
	var progress float64
	if region.AssignmentCount > 0 {
		done := region.DraftCount + region.SubmittedCount + region.ApprovedCount + region.RejectedCount + region.RevokedCount
		progress = float64(done) / float64(region.AssignmentCount) * 100
	}

	return &dto.SurveyRegionResponse{
		ID:              region.ID,
		SurveyID:        region.SurveyID,
		SurveyPeriodID:  region.SurveyPeriodID,
		RegionGroupID:   region.RegionGroupID,
		AssignmentCount: region.AssignmentCount,
		Usaha:           region.Usaha,
		OpenCount:       region.OpenCount,
		DraftCount:      region.DraftCount,
		SubmittedCount:  region.SubmittedCount,
		ApprovedCount:   region.ApprovedCount,
		RejectedCount:   region.RejectedCount,
		RevokedCount:    region.RevokedCount,
		Progress:        progress,
		Level1:          region.Level1,
		Level1Label:     region.Level1Label,
		Level2:          region.Level2,
		Level2Label:     region.Level2Label,
		Level3:          region.Level3,
		Level3Label:     region.Level3Label,
		Level4:          region.Level4,
		Level4Label:     region.Level4Label,
		Level5:          region.Level5,
		Level5Label:     region.Level5Label,
		Level6:          region.Level6,
		Level6Label:     region.Level6Label,
		PJ:              region.PJ,
		PML:             region.PML,
		PPL:             region.PPL,
		FullCode:        region.FullCode,
	}
}

func ToSurveyRegionResponses(regions []models.Region) []dto.SurveyRegionResponse {
	responses := make([]dto.SurveyRegionResponse, 0, len(regions))
	for _, region := range regions {
		responses = append(responses, *ToSurveyRegionResponse(&region))
	}
	return responses
}
