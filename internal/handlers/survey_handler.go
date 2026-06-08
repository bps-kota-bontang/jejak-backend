package handlers

import (
	"fmt"
	"jejak/internal/dto"
	apperrors "jejak/internal/errors"
	"jejak/internal/mappers"
	"jejak/internal/services"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type SurveyHandler struct {
	service  *services.SurveyService
	validate *validator.Validate
}

func NewSurveyHandler(service *services.SurveyService, validate *validator.Validate) *SurveyHandler {
	return &SurveyHandler{
		service:  service,
		validate: validate,
	}
}

func (h *SurveyHandler) GetAll(c fiber.Ctx) error {
	surveys, err := h.service.GetAll()
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, mappers.ToSurveyResponses(surveys), "Surveys retrieved successfully")
}

func (h *SurveyHandler) CreateSurvey(c fiber.Ctx) error {
	var req dto.CreateSurveyRequest
	if err := c.Bind().Body(&req); err != nil {
		return respondValidation(c, err)
	}
	if err := h.validate.Struct(req); err != nil {
		return respondValidation(c, err)
	}

	if err := h.service.CreateSurvey(req); err != nil {
		return respondError(c, err)
	}

	return respondCreated(c, nil, "Survey created successfully")
}

func (h *SurveyHandler) UpdateSurvey(c fiber.Ctx) error {
	surveyPeriodID := c.Params("surveyPeriodId")
	if surveyPeriodID == "" {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "surveyPeriodId is required"))
	}

	var req dto.UpdateSurveyRequest
	if err := c.Bind().Body(&req); err != nil {
		return respondValidation(c, err)
	}
	if err := h.validate.Struct(req); err != nil {
		return respondValidation(c, err)
	}

	if err := h.service.UpdateSurvey(surveyPeriodID, req); err != nil {
		return respondError(c, err)
	}

	return respondOK(c, nil, "Survey updated successfully")
}

func (h *SurveyHandler) SyncSurveyAssignments(c fiber.Ctx) error {
	surveyPeriodID := c.Params("surveyPeriodId")
	if surveyPeriodID == "" {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "surveyPeriodId is required"))
	}

	result, err := h.service.SyncSurveyAssignments(c.Context(), surveyPeriodID)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, result, "Survey assignments synced successfully")
}

func (h *SurveyHandler) GetBySurveyPeriodID(c fiber.Ctx) error {
	surveyPeriodID := c.Params("surveyPeriodId")
	if surveyPeriodID == "" {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "surveyPeriodId is required"))
	}

	survey, err := h.service.GetBySurveyPeriodID(surveyPeriodID)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, mappers.ToSurveyResponse(survey), "Survey retrieved successfully")
}

func (h *SurveyHandler) GetAssignments(c fiber.Ctx) error {
	surveyPeriodID := c.Params("surveyPeriodId")
	if surveyPeriodID == "" {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "surveyPeriodId is required"))
	}

	query := dto.AssignmentRegionFilterQuery{
		RegionFullCode: c.Query("region_full_code"),
		RegionLevel1:   c.Query("region_level_1"),
		RegionLevel2:   c.Query("region_level_2"),
		RegionLevel3:   c.Query("region_level_3"),
		RegionLevel4:   c.Query("region_level_4"),
		RegionLevel5:   c.Query("region_level_5"),
		RegionLevel6:   c.Query("region_level_6"),
	}

	assignments, err := h.service.GetAssignmentsBySurveyPeriodIDWithRegionFilter(surveyPeriodID, query)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, mappers.ToAssignmentResponses(assignments), "Assignments retrieved successfully")
}

func (h *SurveyHandler) GetLogs(c fiber.Ctx) error {
	surveyPeriodID := c.Params("surveyPeriodId")
	if surveyPeriodID == "" {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "surveyPeriodId is required"))
	}

	regionFullCode := strings.TrimSpace(c.Query("region_full_code"))
	if regionFullCode == "" {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "region_full_code is required"))
	}

	actionedAtFrom, err := parseActionedAtFromQuery(c.Query("actioned_at_from"))
	if err != nil {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "actioned_at_from must be RFC3339 or YYYY-MM-DD"))
	}

	actionedAtTo, err := parseActionedAtToQuery(c.Query("actioned_at_to"))
	if err != nil {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "actioned_at_to must be RFC3339 or YYYY-MM-DD"))
	}

	if actionedAtFrom != nil && actionedAtTo != nil && actionedAtFrom.After(*actionedAtTo) {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "actioned_at_from must be before or equal to actioned_at_to"))
	}

	logs, err := h.service.GetLogsBySurveyPeriodIDAndRegionFullCode(
		surveyPeriodID,
		regionFullCode,
		actionedAtFrom,
		actionedAtTo,
	)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, mappers.ToAssignmentLogPointResponses(logs), "Logs retrieved successfully")
}

func parseActionedAtFromQuery(value string) (*time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}

	if t, err := time.Parse(time.RFC3339, trimmed); err == nil {
		return &t, nil
	}

	if t, err := time.Parse("2006-01-02", trimmed); err == nil {
		return &t, nil
	}

	return nil, fmt.Errorf("invalid actioned_at_from format")
}

func parseActionedAtToQuery(value string) (*time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}

	if t, err := time.Parse(time.RFC3339, trimmed); err == nil {
		return &t, nil
	}

	if t, err := time.Parse("2006-01-02", trimmed); err == nil {
		endOfDay := t.Add(24*time.Hour - time.Nanosecond)
		return &endOfDay, nil
	}

	return nil, fmt.Errorf("invalid actioned_at_to format")
}

func (h *SurveyHandler) AnalyzeSurvey(c fiber.Ctx) error {
	surveyPeriodID := c.Params("surveyPeriodId")
	if surveyPeriodID == "" {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "surveyPeriodId is required"))
	}

	result, err := h.service.AnalyzeSurvey(c.Context(), surveyPeriodID)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, result, "Survey analyzed successfully")
}

func (h *SurveyHandler) SyncSurveyRegions(c fiber.Ctx) error {
	surveyPeriodID := c.Params("surveyPeriodId")
	if surveyPeriodID == "" {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "surveyPeriodId is required"))
	}

	var req dto.SyncSurveyRegionsRequest
	if err := c.Bind().Body(&req); err != nil {
		return respondValidation(c, err)
	}
	if err := h.validate.Struct(req); err != nil {
		return respondValidation(c, err)
	}

	result, err := h.service.SyncSurveyRegions(c.Context(), surveyPeriodID, req)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, result, "Survey regions synced successfully")
}

func (h *SurveyHandler) GetRegionMetadata(c fiber.Ctx) error {
	surveyPeriodID := c.Params("surveyPeriodId")
	if surveyPeriodID == "" {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "surveyPeriodId is required"))
	}

	result, err := h.service.GetRegionMetadataBySurveyPeriodID(surveyPeriodID)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, result, "Survey region metadata retrieved successfully")
}

func (h *SurveyHandler) GetRegions(c fiber.Ctx) error {
	surveyPeriodID := c.Params("surveyPeriodId")
	if surveyPeriodID == "" {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "surveyPeriodId is required"))
	}

	query := dto.AssignmentRegionFilterQuery{
		RegionFullCode: c.Query("region_full_code"),
		RegionLevel1:   c.Query("region_level_1"),
		RegionLevel2:   c.Query("region_level_2"),
		RegionLevel3:   c.Query("region_level_3"),
		RegionLevel4:   c.Query("region_level_4"),
		RegionLevel5:   c.Query("region_level_5"),
		RegionLevel6:   c.Query("region_level_6"),
	}

	regions, err := h.service.GetRegionsBySurveyPeriodID(surveyPeriodID, query)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, mappers.ToSurveyRegionResponses(regions), "Survey regions retrieved successfully")
}
