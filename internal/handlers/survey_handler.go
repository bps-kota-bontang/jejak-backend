package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"jejak/internal/dto"
	apperrors "jejak/internal/errors"
	"jejak/internal/mappers"
	"jejak/internal/repositories"
	"jejak/internal/services"
	"jejak/internal/tasks"
	"jejak/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/hibiken/asynq"
)

type SurveyHandler struct {
	service  *services.SurveyService
	validate *validator.Validate
	queue    *asynq.Client
}

func NewSurveyHandler(service *services.SurveyService, validate *validator.Validate, queue *asynq.Client) *SurveyHandler {
	return &SurveyHandler{
		service:  service,
		validate: validate,
		queue:    queue,
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

	roles, _ := c.Locals("roles").([]string)
	if !utils.IsAdmin(roles) {
		return respondError(c, apperrors.NewHttpError(fiber.StatusForbidden, "You are not authorized to sync all assignments"))
	}

	regionTargets, err := h.service.GetSyncRegionTargets(surveyPeriodID)
	if err != nil {
		return respondError(c, err)
	}

	queuedTasks := 0
	alreadyQueued := 0
	for _, regionFullCode := range regionTargets {
		info, err := tasks.EnqueueSurveySyncRegion(c.Context(), h.queue, surveyPeriodID, regionFullCode)
		if err != nil {
			return respondError(c, err)
		}

		if info == nil {
			alreadyQueued++
			continue
		}

		queuedTasks++
	}

	if queuedTasks == 0 {
		return respondAccepted(c, fiber.Map{
			"survey_period_id": surveyPeriodID,
			"total_regions":    len(regionTargets),
			"queued_tasks":     0,
			"already_queued":   alreadyQueued,
		}, "All survey region sync tasks are already queued or running")
	}

	return respondAccepted(c, fiber.Map{
		"survey_period_id": surveyPeriodID,
		"total_regions":    len(regionTargets),
		"queued_tasks":     queuedTasks,
		"already_queued":   alreadyQueued,
	}, "Survey region sync tasks queued successfully")
}

func (h *SurveyHandler) SyncSurveyAssignmentsByRegion(c fiber.Ctx) error {
	surveyPeriodID := c.Params("surveyPeriodId")
	if surveyPeriodID == "" {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "surveyPeriodId is required"))
	}

	regionFullCode := strings.TrimSpace(c.Params("regionFullCode"))
	if regionFullCode == "" {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "regionFullCode is required"))
	}

	info, err := tasks.EnqueueSurveySyncRegion(c.Context(), h.queue, surveyPeriodID, regionFullCode)
	if err != nil {
		return respondError(c, err)
	}

	if info == nil {
		return respondAccepted(c, dto.QueuedSurveyTaskResponse{
			SurveyPeriodID: surveyPeriodID,
			RegionFullCode: regionFullCode,
		}, "Survey region sync already queued or running")
	}

	return respondAccepted(c, dto.QueuedSurveyTaskResponse{
		TaskID:         info.ID,
		Queue:          info.Queue,
		Type:           info.Type,
		SurveyPeriodID: surveyPeriodID,
		RegionFullCode: regionFullCode,
	}, "Survey region sync queued successfully")
}

func (h *SurveyHandler) ImportSurveyRegions(c fiber.Ctx) error {
	surveyPeriodID := c.Params("surveyPeriodId")
	if surveyPeriodID == "" {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "surveyPeriodId is required"))
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "file is required"))
	}

	src, err := fileHeader.Open()
	if err != nil {
		return respondError(c, err)
	}
	defer src.Close()

	raw, err := io.ReadAll(src)
	if err != nil {
		return respondError(c, err)
	}

	result, err := h.service.ImportSurveyRegions(c.Context(), surveyPeriodID, raw)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, result, "Survey regions imported successfully")
}

func (h *SurveyHandler) ImportSurveyAssignments(c fiber.Ctx) error {
	surveyPeriodID := c.Params("surveyPeriodId")
	if surveyPeriodID == "" {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "surveyPeriodId is required"))
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "file is required"))
	}

	src, err := fileHeader.Open()
	if err != nil {
		return respondError(c, err)
	}
	defer src.Close()

	raw, err := io.ReadAll(src)
	if err != nil {
		return respondError(c, err)
	}

	result, err := h.service.ImportSurveyAssignments(c.Context(), surveyPeriodID, raw)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, result, "Survey assignments imported successfully")
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

	roles, _ := c.Locals("roles").([]string)
	if !utils.IsAdmin(roles) {
		return respondError(c, apperrors.NewHttpError(fiber.StatusForbidden, "You are not authorized to analyze all assignments"))
	}

	regionTargets, err := h.service.GetSyncRegionTargets(surveyPeriodID)
	if err != nil {
		return respondError(c, err)
	}

	queuedTasks := 0
	alreadyQueued := 0
	for _, regionFullCode := range regionTargets {
		info, err := tasks.EnqueueSurveyAnalyzeRegion(c.Context(), h.queue, surveyPeriodID, regionFullCode)
		if err != nil {
			log.Printf("[handler][analyze][error] enqueue region analyze failed surveyPeriodID=%s regionFullCode=%s err=%v", surveyPeriodID, regionFullCode, err)
			return respondError(c, apperrors.NewHttpError(http.StatusServiceUnavailable, "failed to enqueue survey region analysis task"))
		}

		if info == nil {
			alreadyQueued++
			continue
		}

		queuedTasks++
	}

	if queuedTasks == 0 {
		return respondAccepted(c, fiber.Map{
			"survey_period_id": surveyPeriodID,
			"total_regions":    len(regionTargets),
			"queued_tasks":     0,
			"already_queued":   alreadyQueued,
		}, "All survey region analysis tasks are already queued or running")
	}

	return respondAccepted(c, fiber.Map{
		"survey_period_id": surveyPeriodID,
		"total_regions":    len(regionTargets),
		"queued_tasks":     queuedTasks,
		"already_queued":   alreadyQueued,
	}, "Survey region analysis tasks queued successfully")
}

func (h *SurveyHandler) AnalyzeSurveyByRegion(c fiber.Ctx) error {
	surveyPeriodID := c.Params("surveyPeriodId")
	if surveyPeriodID == "" {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "surveyPeriodId is required"))
	}

	regionFullCode := strings.TrimSpace(c.Params("regionFullCode"))
	if regionFullCode == "" {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "regionFullCode is required"))
	}

	info, err := tasks.EnqueueSurveyAnalyzeRegion(c.Context(), h.queue, surveyPeriodID, regionFullCode)
	if err != nil {
		log.Printf("[handler][analyze][error] enqueue analyze by region failed surveyPeriodID=%s regionFullCode=%s err=%v", surveyPeriodID, regionFullCode, err)
		return respondError(c, apperrors.NewHttpError(http.StatusServiceUnavailable, "failed to enqueue survey region analysis task"))
	}

	if info == nil {
		return respondAccepted(c, dto.QueuedSurveyTaskResponse{
			SurveyPeriodID: surveyPeriodID,
			RegionFullCode: regionFullCode,
		}, "Survey region analysis already queued or running")
	}

	return respondAccepted(c, dto.QueuedSurveyTaskResponse{
		TaskID:         info.ID,
		Queue:          info.Queue,
		Type:           info.Type,
		SurveyPeriodID: surveyPeriodID,
		RegionFullCode: regionFullCode,
	}, "Survey region analysis queued successfully")
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

	info, err := tasks.EnqueueSurveySyncRegion(c.Context(), h.queue, surveyPeriodID, req.RegionGroupID)
	if err != nil {
		return respondError(c, err)
	}

	if info == nil {
		return respondAccepted(c, dto.QueuedSurveyTaskResponse{
			SurveyPeriodID: surveyPeriodID,
			RegionFullCode: req.RegionGroupID,
		}, "Survey region sync already queued or running")
	}

	return respondAccepted(c, dto.QueuedSurveyTaskResponse{
		TaskID:         info.ID,
		Queue:          info.Queue,
		Type:           info.Type,
		SurveyPeriodID: surveyPeriodID,
		RegionFullCode: req.RegionGroupID,
	}, "Survey region sync queued successfully")
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
		Assignment:     c.Query("assignment_filter"),
		Status:         c.Query("status_filter"),
		Page:           fiber.Query[int](c, "page", 1),
		PerPage:        fiber.Query[int](c, "per_page", 10),
	}

	if query.Page < 1 {
		query.Page = 1
	}
	if query.PerPage < 1 {
		query.PerPage = 10
	}
	if query.PerPage > 1000 {
		query.PerPage = 1000
	}

	regions, total, err := h.service.GetRegionsBySurveyPeriodID(surveyPeriodID, query)
	if err != nil {
		return respondError(c, err)
	}

	meta := utils.NewPaginationMeta(total, query.Page, query.PerPage)
	return respondOKWithMeta(c, mappers.ToSurveyRegionResponses(regions), "Survey regions retrieved successfully", meta)
}

func (h *SurveyHandler) GetRegionFilterOptions(c fiber.Ctx) error {
	surveyPeriodID := c.Params("surveyPeriodId")
	if surveyPeriodID == "" {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "surveyPeriodId is required"))
	}

	// Get filter parameters from query
	filter := repositories.RegionLevelFilter{
		Level1: c.Query("level1"),
		Level2: c.Query("level2"),
		Level3: c.Query("level3"),
		Level4: c.Query("level4"),
		Level5: c.Query("level5"),
	}

	result, err := h.service.GetRegionFilterOptionsWithFilters(surveyPeriodID, filter)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, result, "Region filter options retrieved successfully")
}
