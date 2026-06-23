package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"jejak/internal/dto"
	apperrors "jejak/internal/errors"
	"jejak/internal/models"
	"jejak/internal/repositories"

	"gorm.io/gorm"
)

type SurveyService struct {
	surveyRepo     repositories.SurveyRepository
	assignmentRepo repositories.AssignmentRepository
	logRepo        repositories.LogRepository
	answerRepo     repositories.AnswerRepository
	fasihService   *FasihService
	assignmentSvc  *AssignmentService
}

const maxSyncPageLength = 50

var defaultGMTPlus8 = time.FixedZone("GMT+8", 8*60*60)

type regionNode struct {
	fullCode string
	levels   [6]*string
	labels   [6]*string
}

func NewSurveyService(
	surveyRepo repositories.SurveyRepository,
	assignmentRepo repositories.AssignmentRepository,
	logRepo repositories.LogRepository,
	answerRepo repositories.AnswerRepository,
	fasihService *FasihService,
	assignmentSvc *AssignmentService,
) *SurveyService {
	return &SurveyService{
		surveyRepo:     surveyRepo,
		assignmentRepo: assignmentRepo,
		logRepo:        logRepo,
		answerRepo:     answerRepo,
		fasihService:   fasihService,
		assignmentSvc:  assignmentSvc,
	}
}

func (s *SurveyService) AnalyzeSurvey(ctx context.Context, surveyPeriodID string) (*dto.SurveyFraudAnalysisResult, error) {
	startedAt := time.Now()
	log.Printf("[analyze] start surveyPeriodID=%s", surveyPeriodID)

	assignments, err := s.assignmentRepo.FindBySurveyPeriodID(surveyPeriodID)
	if err != nil {
		log.Printf("[analyze][error] load assignments failed surveyPeriodID=%s err=%v", surveyPeriodID, err)
		return nil, err
	}

	result := &dto.SurveyFraudAnalysisResult{
		SurveyPeriodID:   surveyPeriodID,
		TotalAssignments: len(assignments),
		GeneratedAt:      time.Now(),
		Assignments:      make([]dto.AssignmentSurveyAnalysis, 0, len(assignments)),
	}

	for _, assignment := range assignments {
		select {
		case <-ctx.Done():
			log.Printf("[analyze][error] canceled surveyPeriodID=%s err=%v", surveyPeriodID, ctx.Err())
			return nil, ctx.Err()
		default:
		}

		analysis, err := s.assignmentSvc.AnalyzeAssignment(ctx, assignment.AssignmentID)
		if err != nil {
			log.Printf("[analyze][error] analyze assignment failed surveyPeriodID=%s assignmentID=%s err=%v", surveyPeriodID, assignment.AssignmentID, err)
			return nil, fmt.Errorf("analyze assignment %s: %w", assignment.AssignmentID, err)
		}

		result.AnalyzedAssignments++
		result.Assignments = append(result.Assignments, *analysis)

		if result.AnalyzedAssignments%10 == 0 || result.AnalyzedAssignments == result.TotalAssignments {
			log.Printf("[analyze] progress surveyPeriodID=%s analyzed=%d/%d", surveyPeriodID, result.AnalyzedAssignments, result.TotalAssignments)
		}
	}

	log.Printf("[analyze] completed surveyPeriodID=%s analyzed=%d total=%d duration=%s", surveyPeriodID, result.AnalyzedAssignments, result.TotalAssignments, time.Since(startedAt).Round(time.Second))

	return result, nil
}

func (s *SurveyService) AnalyzeSurveyByRegion(ctx context.Context, surveyPeriodID string, regionFullCode string) (*dto.SurveyFraudAnalysisResult, error) {
	startedAt := time.Now()
	trimmedRegionFullCode := strings.TrimSpace(regionFullCode)
	log.Printf("[analyze] start surveyPeriodID=%s regionFullCode=%s", surveyPeriodID, trimmedRegionFullCode)

	assignments, err := s.assignmentRepo.FindBySurveyPeriodIDWithFilter(surveyPeriodID, repositories.AssignmentRegionFilter{
		RegionFullCode: trimmedRegionFullCode,
	})
	if err != nil {
		log.Printf("[analyze][error] load assignments by region failed surveyPeriodID=%s regionFullCode=%s err=%v", surveyPeriodID, trimmedRegionFullCode, err)
		return nil, err
	}

	result := &dto.SurveyFraudAnalysisResult{
		SurveyPeriodID:   surveyPeriodID,
		TotalAssignments: len(assignments),
		GeneratedAt:      time.Now(),
		Assignments:      make([]dto.AssignmentSurveyAnalysis, 0, len(assignments)),
	}

	for _, assignment := range assignments {
		select {
		case <-ctx.Done():
			log.Printf("[analyze][error] canceled surveyPeriodID=%s regionFullCode=%s err=%v", surveyPeriodID, trimmedRegionFullCode, ctx.Err())
			return nil, ctx.Err()
		default:
		}

		analysis, err := s.assignmentSvc.AnalyzeAssignment(ctx, assignment.AssignmentID)
		if err != nil {
			log.Printf("[analyze][error] analyze assignment failed surveyPeriodID=%s regionFullCode=%s assignmentID=%s err=%v", surveyPeriodID, trimmedRegionFullCode, assignment.AssignmentID, err)
			return nil, fmt.Errorf("analyze assignment %s: %w", assignment.AssignmentID, err)
		}

		result.AnalyzedAssignments++
		result.Assignments = append(result.Assignments, *analysis)

		if result.AnalyzedAssignments%10 == 0 || result.AnalyzedAssignments == result.TotalAssignments {
			log.Printf("[analyze] progress surveyPeriodID=%s regionFullCode=%s analyzed=%d/%d", surveyPeriodID, trimmedRegionFullCode, result.AnalyzedAssignments, result.TotalAssignments)
		}
	}

	log.Printf("[analyze] completed surveyPeriodID=%s regionFullCode=%s analyzed=%d total=%d duration=%s", surveyPeriodID, trimmedRegionFullCode, result.AnalyzedAssignments, result.TotalAssignments, time.Since(startedAt).Round(time.Second))

	return result, nil
}

func (s *SurveyService) CreateSurvey(req dto.CreateSurveyRequest) error {
	name := strings.TrimSpace(req.Name)
	surveyID := strings.TrimSpace(req.SurveyID)
	surveyPeriodID := strings.TrimSpace(req.SurveyPeriodID)
	xsrfToken := strings.TrimSpace(req.XSRFToken)
	cookie := strings.TrimSpace(req.Cookie)
	regionLevel1 := strings.TrimSpace(req.RegionLevel1)
	regionLevel2 := strings.TrimSpace(req.RegionLevel2)
	areaID := strings.TrimSpace(req.AreaID)
	geoJSONKey := strings.TrimSpace(req.GeoJSONKey)

	if name == "" || surveyID == "" || surveyPeriodID == "" || xsrfToken == "" || cookie == "" || regionLevel1 == "" || regionLevel2 == "" || areaID == "" || geoJSONKey == "" {
		return apperrors.NewHttpError(http.StatusBadRequest, "name, survey_id, survey_period_id, xsrf_token, cookie, region_level_1, region_level_2, area_id, dan geojson_key wajib diisi")
	}

	logDateFrom, err := parseSurveyDatePtr(req.LogDateFrom)
	if err != nil {
		return apperrors.NewHttpError(http.StatusBadRequest, "log_date_from harus format YYYY-MM-DD")
	}

	logDateTo, err := parseSurveyDatePtr(req.LogDateTo)
	if err != nil {
		return apperrors.NewHttpError(http.StatusBadRequest, "log_date_to harus format YYYY-MM-DD")
	}

	if logDateFrom != nil && logDateTo != nil && logDateFrom.After(*logDateTo) {
		return apperrors.NewHttpError(http.StatusBadRequest, "log_date_from tidak boleh lebih besar dari log_date_to")
	}

	survey := &models.Survey{
		Name:            name,
		SurveyID:        surveyID,
		SurveyPeriodID:  surveyPeriodID,
		XSRFToken:       xsrfToken,
		Cookie:          cookie,
		RegionLevel1:    regionLevel1,
		RegionLevel2:    regionLevel2,
		LogDeltaMaxMins: req.LogDeltaMaxMins,
		LogDateFrom:     logDateFrom,
		LogDateTo:       logDateTo,
		AreaID:          areaID,
		GeoJSONKey:      geoJSONKey,
	}

	return s.surveyRepo.Upsert(survey)
}

func (s *SurveyService) UpdateSurvey(surveyPeriodID string, req dto.UpdateSurveyRequest) error {
	name := strings.TrimSpace(req.Name)
	regionLevel1 := strings.TrimSpace(req.RegionLevel1)
	regionLevel2 := strings.TrimSpace(req.RegionLevel2)
	areaID := strings.TrimSpace(req.AreaID)
	geoJSONKey := strings.TrimSpace(req.GeoJSONKey)
	surveyID := strings.TrimSpace(req.SurveyID)
	xsrfToken := strings.TrimSpace(req.XSRFToken)
	cookie := strings.TrimSpace(req.Cookie)

	if name == "" || surveyID == "" || xsrfToken == "" || cookie == "" || regionLevel1 == "" || regionLevel2 == "" || areaID == "" || geoJSONKey == "" {
		return apperrors.NewHttpError(http.StatusBadRequest, "name, survey_id, xsrf_token, cookie, region_level_1, region_level_2, area_id, dan geojson_key wajib diisi")
	}

	existing, err := s.surveyRepo.FindBySurveyPeriodID(surveyPeriodID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperrors.NewHttpError(http.StatusNotFound, "survey tidak ditemukan")
		}
		return err
	}

	if surveyID != existing.SurveyID ||
		regionLevel1 != existing.RegionLevel1 ||
		regionLevel2 != existing.RegionLevel2 ||
		areaID != existing.AreaID ||
		geoJSONKey != existing.GeoJSONKey {
		return apperrors.NewHttpError(http.StatusBadRequest, "kode survey, region level 1-2, area, dan geojson key tidak boleh diubah setelah survey dibuat")
	}

	logDateFrom, err := parseSurveyDatePtr(req.LogDateFrom)
	if err != nil {
		return apperrors.NewHttpError(http.StatusBadRequest, "log_date_from harus format YYYY-MM-DD")
	}

	logDateTo, err := parseSurveyDatePtr(req.LogDateTo)
	if err != nil {
		return apperrors.NewHttpError(http.StatusBadRequest, "log_date_to harus format YYYY-MM-DD")
	}

	if logDateFrom != nil && logDateTo != nil && logDateFrom.After(*logDateTo) {
		return apperrors.NewHttpError(http.StatusBadRequest, "log_date_from tidak boleh lebih besar dari log_date_to")
	}

	survey := &models.Survey{
		Name:            name,
		SurveyID:        existing.SurveyID,
		SurveyPeriodID:  surveyPeriodID,
		XSRFToken:       xsrfToken,
		Cookie:          cookie,
		RegionLevel1:    existing.RegionLevel1,
		RegionLevel2:    existing.RegionLevel2,
		LogDeltaMaxMins: req.LogDeltaMaxMins,
		LogDateFrom:     logDateFrom,
		LogDateTo:       logDateTo,
		AreaID:          existing.AreaID,
		GeoJSONKey:      existing.GeoJSONKey,
	}

	return s.surveyRepo.UpdateBySurveyPeriodID(surveyPeriodID, survey)
}

func (s *SurveyService) GetAll() ([]models.Survey, error) {
	return s.surveyRepo.FindAll()
}

func (s *SurveyService) GetBySurveyPeriodID(surveyPeriodID string) (*models.Survey, error) {
	return s.surveyRepo.FindBySurveyPeriodID(surveyPeriodID)
}

func (s *SurveyService) GetAssignmentsBySurveyPeriodID(surveyPeriodID string) ([]models.Assignment, error) {
	return s.assignmentRepo.FindBySurveyPeriodID(surveyPeriodID)
}

func (s *SurveyService) GetAssignmentsBySurveyPeriodIDWithRegionFilter(surveyPeriodID string, query dto.AssignmentRegionFilterQuery) ([]models.Assignment, error) {
	filter := repositories.AssignmentRegionFilter{
		RegionFullCode: query.RegionFullCode,
		RegionLevel1:   query.RegionLevel1,
		RegionLevel2:   query.RegionLevel2,
		RegionLevel3:   query.RegionLevel3,
		RegionLevel4:   query.RegionLevel4,
		RegionLevel5:   query.RegionLevel5,
		RegionLevel6:   query.RegionLevel6,
	}

	return s.assignmentRepo.FindBySurveyPeriodIDWithFilter(surveyPeriodID, filter)
}

func (s *SurveyService) GetLogsBySurveyPeriodIDAndRegionFullCode(
	surveyPeriodID string,
	regionFullCode string,
	actionedAtFrom *time.Time,
	actionedAtTo *time.Time,
) ([]models.Log, error) {
	logs, err := s.logRepo.FindBySurveyPeriodIDRegionFullCodeAndActionedAt(
		surveyPeriodID,
		regionFullCode,
		actionedAtFrom,
		actionedAtTo,
	)
	if err != nil {
		return nil, err
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].ActionedAt.Before(logs[j].ActionedAt)
	})

	return logs, nil
}

func (s *SurveyService) GetRegionMetadataBySurveyPeriodID(surveyPeriodID string) (*dto.SurveyRegionMetadataResponse, error) {
	survey, err := s.surveyRepo.FindBySurveyPeriodID(surveyPeriodID)
	if err != nil {
		return nil, err
	}

	if survey.RegionGroupID == nil {
		return nil, fmt.Errorf("region metadata has not been synced for survey_period_id %s", surveyPeriodID)
	}

	creds := dto.FasihCredentials{
		Cookie:    survey.Cookie,
		XSRFToken: survey.XSRFToken,
	}

	metadataResp, err := s.fasihService.GetRegionMetadata(context.Background(), creds, dto.FasihRegionMetadataRequest{GroupID: *survey.RegionGroupID})
	if err != nil {
		return nil, err
	}

	metadataLevels := make([]dto.SurveyRegionMetadataLevelResponse, 0, len(metadataResp.Data.Level))
	for _, level := range metadataResp.Data.Level {
		metadataLevels = append(metadataLevels, dto.SurveyRegionMetadataLevelResponse{
			ID:   level.ID,
			Name: level.Name,
		})
	}

	return &dto.SurveyRegionMetadataResponse{
		RegionGroupID:       *survey.RegionGroupID,
		LevelCount:          metadataResp.Data.LevelCount,
		SmallestRegionLevel: metadataResp.Data.SmallestRegionLevel,
		GroupName:           metadataResp.Data.GroupName,
		IsActive:            metadataResp.Data.IsActive,
		IsPublic:            metadataResp.Data.IsPublic,
		Level:               metadataLevels,
	}, nil
}

func (s *SurveyService) GetRegionsBySurveyPeriodID(surveyPeriodID string, query dto.AssignmentRegionFilterQuery) ([]models.Region, int64, error) {
	page := query.Page
	if page < 1 {
		page = 1
	}

	perPage := query.PerPage
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 1000 {
		perPage = 1000
	}

	filter := repositories.AssignmentRegionFilter{
		RegionFullCode: query.RegionFullCode,
		RegionLevel1:   query.RegionLevel1,
		RegionLevel2:   query.RegionLevel2,
		RegionLevel3:   query.RegionLevel3,
		RegionLevel4:   query.RegionLevel4,
		RegionLevel5:   query.RegionLevel5,
		RegionLevel6:   query.RegionLevel6,
		Assignment:     query.Assignment,
		Status:         query.Status,
	}

	total, err := s.surveyRepo.CountBySurveyPeriodIDWithFilter(surveyPeriodID, filter)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	regions, err := s.surveyRepo.FindBySurveyPeriodIDWithFilterPaginated(surveyPeriodID, filter, perPage, offset)
	if err != nil {
		return nil, 0, err
	}

	return regions, total, nil
}

func (s *SurveyService) GetRegionFilterOptions(surveyPeriodID string) (*dto.RegionFilterOptionsResponse, error) {
	return s.GetRegionFilterOptionsWithFilters(surveyPeriodID, repositories.RegionLevelFilter{})
}

func (s *SurveyService) GetRegionFilterOptionsWithFilters(surveyPeriodID string, filter repositories.RegionLevelFilter) (*dto.RegionFilterOptionsResponse, error) {
	// Calculate response depth: how many levels to include in response
	// based on number of filters provided
	filterCount := 0
	if filter.Level1 != "" {
		filterCount++
	}
	if filter.Level2 != "" {
		filterCount++
	}
	if filter.Level3 != "" {
		filterCount++
	}
	if filter.Level4 != "" {
		filterCount++
	}
	if filter.Level5 != "" {
		filterCount++
	}

	// Depth = max(3, filterCount + 1), capped at 6
	// So: 0 filters → 3 levels, 1 filter → 3 levels, 2 filters → 3 levels, 3 filters → 4 levels, etc.
	depth := filterCount + 1
	if depth < 3 {
		depth = 3
	}
	if depth > 6 {
		depth = 6
	}

	// Only fetch the levels we need based on depth
	toOptions := func(options []repositories.RegionLevelOption) []dto.RegionFilterOption {
		dtoOptions := make([]dto.RegionFilterOption, len(options))
		for i, opt := range options {
			label := opt.Label
			if label == "" {
				label = opt.Value
			}
			dtoOptions[i] = dto.RegionFilterOption{
				Value: opt.Value,
				Label: label,
			}
		}
		return dtoOptions
	}

	response := &dto.RegionFilterOptionsResponse{
		Level1: make([]dto.RegionFilterOption, 0),
		Level2: make([]dto.RegionFilterOption, 0),
		Level3: make([]dto.RegionFilterOption, 0),
		Level4: make([]dto.RegionFilterOption, 0),
		Level5: make([]dto.RegionFilterOption, 0),
		Level6: make([]dto.RegionFilterOption, 0),
	}

	if depth >= 1 {
		level1, err := s.surveyRepo.GetDistinctRegionLevel1(surveyPeriodID)
		if err != nil {
			return nil, err
		}
		response.Level1 = toOptions(level1)
	}

	if depth >= 2 {
		level2, err := s.surveyRepo.GetDistinctRegionLevel2(surveyPeriodID, repositories.RegionLevelFilter{Level1: filter.Level1})
		if err != nil {
			return nil, err
		}
		response.Level2 = toOptions(level2)
	}

	if depth >= 3 {
		level3, err := s.surveyRepo.GetDistinctRegionLevel3(surveyPeriodID, repositories.RegionLevelFilter{Level1: filter.Level1, Level2: filter.Level2})
		if err != nil {
			return nil, err
		}
		response.Level3 = toOptions(level3)
	}

	if depth >= 4 {
		level4, err := s.surveyRepo.GetDistinctRegionLevel4(surveyPeriodID, repositories.RegionLevelFilter{Level1: filter.Level1, Level2: filter.Level2, Level3: filter.Level3})
		if err != nil {
			return nil, err
		}
		response.Level4 = toOptions(level4)
	}

	if depth >= 5 {
		level5, err := s.surveyRepo.GetDistinctRegionLevel5(surveyPeriodID, repositories.RegionLevelFilter{Level1: filter.Level1, Level2: filter.Level2, Level3: filter.Level3, Level4: filter.Level4})
		if err != nil {
			return nil, err
		}
		response.Level5 = toOptions(level5)
	}

	if depth >= 6 {
		level6, err := s.surveyRepo.GetDistinctRegionLevel6(surveyPeriodID, repositories.RegionLevelFilter{Level1: filter.Level1, Level2: filter.Level2, Level3: filter.Level3, Level4: filter.Level4, Level5: filter.Level5})
		if err != nil {
			return nil, err
		}
		response.Level6 = toOptions(level6)
	}

	return response, nil
}

func (s *SurveyService) ImportSurveyRegions(ctx context.Context, surveyPeriodID string, raw []byte) (*dto.SyncSurveyRegionsResponse, error) {
	_ = ctx

	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, apperrors.NewHttpError(http.StatusBadRequest, "file import region kosong")
	}

	survey, err := s.surveyRepo.FindBySurveyPeriodID(surveyPeriodID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewHttpError(http.StatusNotFound, "survey tidak ditemukan")
		}
		return nil, err
	}

	var payload dto.RegionImportPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, apperrors.NewHttpError(http.StatusBadRequest, "format file region tidak valid")
	}

	if payloadType := strings.TrimSpace(payload.Type); payloadType != "" && payloadType != "region_sync_export_v1" {
		return nil, apperrors.NewHttpError(http.StatusBadRequest, "tipe file region tidak didukung")
	}

	if fileSurveyPeriodID := strings.TrimSpace(payload.SurveyPeriodID); fileSurveyPeriodID != "" && fileSurveyPeriodID != surveyPeriodID {
		return nil, apperrors.NewHttpError(http.StatusBadRequest, "survey_period_id pada file region tidak sesuai")
	}

	groupID := strings.TrimSpace(payload.RegionGroupID)
	if groupID == "" && survey.RegionGroupID != nil {
		groupID = strings.TrimSpace(*survey.RegionGroupID)
	}
	if groupID == "" {
		return nil, apperrors.NewHttpError(http.StatusBadRequest, "region_group_id pada file region wajib diisi")
	}

	regions := make([]models.Region, 0, len(payload.Regions))
	for _, item := range payload.Regions {
		fullCode := strings.TrimSpace(item.FullCode)
		if fullCode == "" {
			continue
		}

		itemSurveyPeriodID := strings.TrimSpace(item.SurveyPeriodID)
		if itemSurveyPeriodID != "" && itemSurveyPeriodID != surveyPeriodID {
			continue
		}

		itemSurveyID := strings.TrimSpace(item.SurveyID)
		if itemSurveyID == "" {
			itemSurveyID = survey.SurveyID
		}

		itemGroupID := strings.TrimSpace(item.RegionGroupID)
		if itemGroupID == "" {
			itemGroupID = groupID
		}

		regions = append(regions, models.Region{
			SurveyID:       itemSurveyID,
			SurveyPeriodID: surveyPeriodID,
			RegionGroupID:  itemGroupID,
			Level1:         trimmedPtr(item.Level1),
			Level1Label:    trimmedPtr(item.Level1Label),
			Level2:         trimmedPtr(item.Level2),
			Level2Label:    trimmedPtr(item.Level2Label),
			Level3:         trimmedPtr(item.Level3),
			Level3Label:    trimmedPtr(item.Level3Label),
			Level4:         trimmedPtr(item.Level4),
			Level4Label:    trimmedPtr(item.Level4Label),
			Level5:         trimmedPtr(item.Level5),
			Level5Label:    trimmedPtr(item.Level5Label),
			Level6:         trimmedPtr(item.Level6),
			Level6Label:    trimmedPtr(item.Level6Label),
			FullCode:       fullCode,
		})
	}

	if len(regions) == 0 {
		return nil, apperrors.NewHttpError(http.StatusBadRequest, "tidak ada data region valid di file import")
	}

	levelCount := payload.LevelCount
	if levelCount <= 0 {
		levelCount = inferRegionLevelCount(regions)
	}
	if levelCount <= 0 {
		levelCount = 1
	}

	if err := s.surveyRepo.UpdateRegionMetadata(surveyPeriodID, groupID, levelCount); err != nil {
		return nil, err
	}

	if err := s.surveyRepo.ReplaceSurveyRegions(surveyPeriodID, regions); err != nil {
		return nil, err
	}

	if err := s.surveyRepo.UpdateSurveyRegionAssignmentCounts(surveyPeriodID); err != nil {
		return nil, err
	}

	return &dto.SyncSurveyRegionsResponse{
		RegionGroupID: groupID,
		LevelCount:    levelCount,
		SavedRegions:  len(regions),
	}, nil
}

func (s *SurveyService) ImportSurveyAssignments(ctx context.Context, surveyPeriodID string, raw []byte) (*dto.SyncSurveyAssignmentsResponse, error) {
	_ = ctx

	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, apperrors.NewHttpError(http.StatusBadRequest, "file import assignment kosong")
	}

	if _, err := s.surveyRepo.FindBySurveyPeriodID(surveyPeriodID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewHttpError(http.StatusNotFound, "survey tidak ditemukan")
		}
		return nil, err
	}

	var payload dto.AssignmentImportPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, apperrors.NewHttpError(http.StatusBadRequest, "format file assignment tidak valid")
	}

	if payloadType := strings.TrimSpace(payload.Type); payloadType != "" && payloadType != "assignment_sync_export_v1" {
		return nil, apperrors.NewHttpError(http.StatusBadRequest, "tipe file assignment tidak didukung")
	}

	if fileSurveyPeriodID := strings.TrimSpace(payload.SurveyPeriodID); fileSurveyPeriodID != "" && fileSurveyPeriodID != surveyPeriodID {
		return nil, apperrors.NewHttpError(http.StatusBadRequest, "survey_period_id pada file assignment tidak sesuai")
	}

	logsByAssignment := make(map[string][]models.Log)
	seenLogHashesByAssignment := make(map[string]map[string]struct{})
	for _, item := range payload.Logs {
		assignmentID := strings.TrimSpace(item.AssignmentID)
		if assignmentID == "" {
			continue
		}

		actionedAt, ok, err := parseOptionalFlexibleTime(item.ActionedAt)
		if err != nil {
			return nil, apperrors.NewHttpError(http.StatusBadRequest, fmt.Sprintf("format actioned_at tidak valid untuk assignment %s", assignmentID))
		}
		if !ok {
			continue
		}

		hash := models.BuildLogEventHash(assignmentID, strings.TrimSpace(item.Action), item.Latitude, item.Longitude, actionedAt)
		if _, ok := seenLogHashesByAssignment[assignmentID]; !ok {
			seenLogHashesByAssignment[assignmentID] = make(map[string]struct{})
		}
		if _, exists := seenLogHashesByAssignment[assignmentID][hash]; exists {
			continue
		}
		seenLogHashesByAssignment[assignmentID][hash] = struct{}{}

		logsByAssignment[assignmentID] = append(logsByAssignment[assignmentID], models.Log{
			AssignmentID: assignmentID,
			EventHash:    hash,
			Action:       strings.TrimSpace(item.Action),
			Latitude:     item.Latitude,
			Longitude:    item.Longitude,
			ActionedAt:   actionedAt,
		})
	}

	answersByAssignment := make(map[string][]models.Answer)
	for _, item := range payload.Answers {
		assignmentID := strings.TrimSpace(item.AssignmentID)
		name := strings.TrimSpace(item.Name)
		if assignmentID == "" || name == "" {
			continue
		}

		answeredAt, hasAnsweredAt, err := parseOptionalFlexibleTime(item.AnsweredAt)
		if err != nil {
			return nil, apperrors.NewHttpError(http.StatusBadRequest, fmt.Sprintf("format answered_at tidak valid untuk assignment %s", assignmentID))
		}
		revisedAt, hasRevisedAt, err := parseOptionalFlexibleTime(item.RevisedAt)
		if err != nil {
			return nil, apperrors.NewHttpError(http.StatusBadRequest, fmt.Sprintf("format revised_at tidak valid untuk assignment %s", assignmentID))
		}

		if !hasAnsweredAt && !hasRevisedAt {
			continue
		}
		if !hasAnsweredAt {
			answeredAt = revisedAt
		}
		if !hasRevisedAt {
			revisedAt = answeredAt
		}

		answersByAssignment[assignmentID] = append(answersByAssignment[assignmentID], models.Answer{
			AssignmentID: assignmentID,
			Name:         name,
			AnsweredAt:   answeredAt,
			RevisedAt:    revisedAt,
		})
	}

	result := &dto.SyncSurveyAssignmentsResponse{
		TotalAssignments: payload.TotalHit,
	}

	for _, item := range payload.Assignments {
		assignmentID := strings.TrimSpace(item.AssignmentID)
		if assignmentID == "" {
			continue
		}

		itemSurveyPeriodID := strings.TrimSpace(item.SurveyPeriodID)
		if itemSurveyPeriodID != "" && itemSurveyPeriodID != surveyPeriodID {
			continue
		}

		submittedAt, hasSubmittedAt, err := parseOptionalFlexibleTime(item.SubmittedAt)
		if err != nil || !hasSubmittedAt {
			return nil, apperrors.NewHttpError(http.StatusBadRequest, fmt.Sprintf("format submitted_at tidak valid untuk assignment %s", assignmentID))
		}

		revisedAt, hasRevisedAt, err := parseOptionalFlexibleTime(item.RevisedAt)
		if err != nil {
			return nil, apperrors.NewHttpError(http.StatusBadRequest, fmt.Sprintf("format revised_at tidak valid untuk assignment %s", assignmentID))
		}
		if !hasRevisedAt {
			revisedAt = submittedAt
		}

		openedAt, hasOpenedAt, err := parseOptionalFlexibleTime(item.OpenedAt)
		if err != nil {
			return nil, apperrors.NewHttpError(http.StatusBadRequest, fmt.Sprintf("format opened_at tidak valid untuk assignment %s", assignmentID))
		}

		startedAt, hasStartedAt, err := parseOptionalFlexibleTime(item.StartedAt)
		if err != nil {
			return nil, apperrors.NewHttpError(http.StatusBadRequest, fmt.Sprintf("format started_at tidak valid untuk assignment %s", assignmentID))
		}

		assignment := &models.Assignment{
			SurveyPeriodID: surveyPeriodID,
			AssignmentID:   assignmentID,
			Status:         models.AssignmentStatusCodeFromInt(item.Status),
			RegionFullCode: trimmedPtr(item.RegionFullCode),
			RegionLevel1:   trimmedPtr(item.RegionLevel1),
			RegionLevel2:   trimmedPtr(item.RegionLevel2),
			RegionLevel3:   trimmedPtr(item.RegionLevel3),
			RegionLevel4:   trimmedPtr(item.RegionLevel4),
			RegionLevel5:   trimmedPtr(item.RegionLevel5),
			RegionLevel6:   trimmedPtr(item.RegionLevel6),
			Latitude:       item.Latitude,
			Longitude:      item.Longitude,
			OpenedAt:       sql.NullTime{Time: openedAt, Valid: hasOpenedAt},
			StartedAt:      sql.NullTime{Time: startedAt, Valid: hasStartedAt},
			SubmittedAt:    submittedAt,
			RevisedAt:      revisedAt,
		}

		if err := s.assignmentRepo.Upsert(assignment); err != nil {
			return nil, err
		}

		result.SavedAssignments++

		logs := logsByAssignment[assignmentID]
		if err := s.logRepo.ReplaceByAssignmentID(assignmentID, logs); err != nil {
			return nil, err
		}
		result.SavedLogs += len(logs)

		answers := answersByAssignment[assignmentID]
		if err := s.answerRepo.ReplaceByAssignmentID(assignmentID, answers); err != nil {
			return nil, err
		}
		result.SavedAnswers += len(answers)
	}

	if result.TotalAssignments == 0 {
		result.TotalAssignments = result.SavedAssignments
	}

	if err := s.surveyRepo.UpdateSurveyRegionAssignmentCounts(surveyPeriodID); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *SurveyService) SyncSurveyRegions(ctx context.Context, surveyPeriodID string, req dto.SyncSurveyRegionsRequest) (*dto.SyncSurveyRegionsResponse, error) {
	if !s.fasihService.IsAvailable(ctx) {
		return nil, apperrors.NewHttpError(http.StatusServiceUnavailable, "Fasih belum dapat diakses saat ini")
	}

	startedAt := time.Now()
	log.Printf("[sync-region] start: surveyPeriodID=%s requestedGroupID=%s", surveyPeriodID, strings.TrimSpace(req.RegionGroupID))

	survey, err := s.surveyRepo.FindBySurveyPeriodID(surveyPeriodID)
	if err != nil {
		log.Printf("[sync-region] failed load survey: surveyPeriodID=%s err=%v", surveyPeriodID, err)
		return nil, err
	}

	groupID := strings.TrimSpace(req.RegionGroupID)
	creds := dto.FasihCredentials{
		Cookie:    survey.Cookie,
		XSRFToken: survey.XSRFToken,
	}

	if groupID == "" {
		log.Printf("[sync-region] resolving region_group_id from survey_id=%s", survey.SurveyID)
		surveyResp, err := s.fasihService.GetSurveyByID(ctx, creds, dto.FasihSurveyByIDRequest{SurveyID: survey.SurveyID})
		if err != nil {
			log.Printf("[sync-region] failed resolve region_group_id: surveyID=%s err=%v", survey.SurveyID, err)
			return nil, err
		}

		groupID = strings.TrimSpace(surveyResp.Data.RegionGroupID)
		if groupID == "" {
			return nil, fmt.Errorf("region_group_id not found for survey_id %s", survey.SurveyID)
		}
	}

	metadataResp, err := s.fasihService.GetRegionMetadata(ctx, creds, dto.FasihRegionMetadataRequest{GroupID: groupID})
	if err != nil {
		log.Printf("[sync-region] failed fetch metadata: surveyPeriodID=%s groupID=%s err=%v", surveyPeriodID, groupID, err)
		return nil, err
	}

	metadata := metadataResp.Data
	if metadata.LevelCount < 1 {
		log.Printf("[sync-region] invalid metadata level count: surveyPeriodID=%s groupID=%s levelCount=%d", surveyPeriodID, groupID, metadata.LevelCount)
		return nil, fmt.Errorf("invalid level count in region metadata")
	}

	log.Printf("[sync-region] metadata loaded: surveyPeriodID=%s groupID=%s levelCount=%d", surveyPeriodID, groupID, metadata.LevelCount)

	if err := s.surveyRepo.UpdateRegionMetadata(surveyPeriodID, groupID, metadata.LevelCount); err != nil {
		log.Printf("[sync-region] failed update survey region metadata: surveyPeriodID=%s groupID=%s err=%v", surveyPeriodID, groupID, err)
		return nil, err
	}

	regions := make([]models.Region, 0)
	parents := []regionNode{{fullCode: ""}}
	startLevel := 1
	scopeLevel1 := strings.TrimSpace(survey.RegionLevel1)
	scopeLevel2 := strings.TrimSpace(survey.RegionLevel2)

	if scopeLevel2 != "" && scopeLevel1 == "" {
		return nil, apperrors.NewHttpError(http.StatusBadRequest, "region_level_1 wajib diisi saat region_level_2 digunakan")
	}

	if scopeLevel1 != "" {
		baseLevels := [6]*string{}
		baseLabels := [6]*string{}
		level1Copy := scopeLevel1
		baseLevels[0] = &level1Copy

		listResp, err := s.fasihService.GetRegionsByLevel(ctx, creds, dto.FasihRegionListRequest{
			GroupID:        groupID,
			Level:          1,
			ParentFullCode: "",
		})
		if err != nil {
			log.Printf("[sync-region] failed fetch scope level1 label: surveyPeriodID=%s regionLevel1=%s err=%v", surveyPeriodID, scopeLevel1, err)
			return nil, err
		}

		for _, item := range listResp.Data {
			if shouldSkipRegionItem(item) {
				continue
			}
			if strings.TrimSpace(item.Code) == scopeLevel1 {
				baseLabels = appendLevelLabel(baseLabels, 1, item.Name)
				break
			}
		}

		parents = []regionNode{{
			fullCode: scopeLevel1,
			levels:   baseLevels,
			labels:   baseLabels,
		}}
		startLevel = 2
		log.Printf("[sync-region] scope enabled: surveyPeriodID=%s regionLevel1=%s regionLevel2=%s", surveyPeriodID, scopeLevel1, scopeLevel2)
	}

	for level := startLevel; level <= metadata.LevelCount; level++ {
		levelSavedStart := len(regions)
		requestCount := 0
		processedCount := 0
		nextParents := make([]regionNode, 0)
		persistLevel := level == metadata.LevelCount

		for _, parent := range parents {
			requestCount++
			listResp, err := s.fasihService.GetRegionsByLevel(ctx, creds, dto.FasihRegionListRequest{
				GroupID:        groupID,
				Level:          level,
				ParentFullCode: parent.fullCode,
			})
			if err != nil {
				log.Printf("[sync-region] failed fetch regions by level: surveyPeriodID=%s groupID=%s level=%d parentFullCode=%s err=%v", surveyPeriodID, groupID, level, parent.fullCode, err)
				return nil, err
			}

			if level == 2 && scopeLevel2 != "" {
				var selected *dto.FasihRegionItem
				for i := range listResp.Data {
					if shouldSkipRegionItem(listResp.Data[i]) {
						continue
					}
					if strings.TrimSpace(listResp.Data[i].Code) == scopeLevel2 {
						selected = &listResp.Data[i]
						break
					}
				}

				if selected == nil {
					log.Printf("[sync-region] level2 scope not found: surveyPeriodID=%s parentFullCode=%s regionLevel2=%s", surveyPeriodID, parent.fullCode, scopeLevel2)
					continue
				}

				levelCodes := appendLevelCode(parent.levels, level, selected.Code)
				levelNames := appendLevelLabel(parent.labels, level, selected.Name)
				region := models.Region{
					SurveyID:       survey.SurveyID,
					SurveyPeriodID: survey.SurveyPeriodID,
					RegionGroupID:  groupID,
					Level1:         levelCodes[0],
					Level1Label:    levelNames[0],
					Level2:         levelCodes[1],
					Level2Label:    levelNames[1],
					Level3:         levelCodes[2],
					Level3Label:    levelNames[2],
					Level4:         levelCodes[3],
					Level4Label:    levelNames[3],
					Level5:         levelCodes[4],
					Level5Label:    levelNames[4],
					Level6:         levelCodes[5],
					Level6Label:    levelNames[5],
					FullCode:       selected.FullCode,
				}

				processedCount++
				if persistLevel {
					regions = append(regions, region)
				}
				nextParents = append(nextParents, regionNode{
					fullCode: selected.FullCode,
					levels:   levelCodes,
					labels:   levelNames,
				})
				continue
			}

			for _, item := range listResp.Data {
				if shouldSkipRegionItem(item) {
					continue
				}

				levelCodes := appendLevelCode(parent.levels, level, item.Code)
				levelNames := appendLevelLabel(parent.labels, level, item.Name)
				region := models.Region{
					SurveyID:       survey.SurveyID,
					SurveyPeriodID: survey.SurveyPeriodID,
					RegionGroupID:  groupID,
					Level1:         levelCodes[0],
					Level1Label:    levelNames[0],
					Level2:         levelCodes[1],
					Level2Label:    levelNames[1],
					Level3:         levelCodes[2],
					Level3Label:    levelNames[2],
					Level4:         levelCodes[3],
					Level4Label:    levelNames[3],
					Level5:         levelCodes[4],
					Level5Label:    levelNames[4],
					Level6:         levelCodes[5],
					Level6Label:    levelNames[5],
					FullCode:       item.FullCode,
				}

				processedCount++
				if persistLevel {
					regions = append(regions, region)
				}
				nextParents = append(nextParents, regionNode{
					fullCode: item.FullCode,
					levels:   levelCodes,
					labels:   levelNames,
				})
			}
		}

		parents = dedupeRegionNodes(nextParents)
		log.Printf("[sync-region] level=%d completed: requests=%d processed=%d saved=%d nextParents=%d totalSaved=%d", level, requestCount, processedCount, len(regions)-levelSavedStart, len(parents), len(regions))
	}

	if err := s.surveyRepo.ReplaceSurveyRegions(surveyPeriodID, regions); err != nil {
		log.Printf("[sync-region] failed replace regions: surveyPeriodID=%s total=%d err=%v", surveyPeriodID, len(regions), err)
		return nil, err
	}

	log.Printf("[sync-region] completed in %s: surveyPeriodID=%s groupID=%s levelCount=%d saved=%d", time.Since(startedAt).Round(time.Second), surveyPeriodID, groupID, metadata.LevelCount, len(regions))

	return &dto.SyncSurveyRegionsResponse{
		RegionGroupID: groupID,
		LevelCount:    metadata.LevelCount,
		SavedRegions:  len(regions),
	}, nil
}

func (s *SurveyService) SyncSurveyAssignments(ctx context.Context, surveyPeriodID string, req dto.SyncSurveyAssignmentsRequest) (*dto.SyncSurveyAssignmentsResponse, error) {
	if !s.fasihService.IsAvailable(ctx) {
		return nil, apperrors.NewHttpError(http.StatusServiceUnavailable, "Fasih belum dapat diakses saat ini")
	}

	requestedRegionFullCode := strings.TrimSpace(req.RegionFullCode)
	if requestedRegionFullCode == "" {
		return nil, apperrors.NewHttpError(http.StatusBadRequest, "region_full_code wajib diisi")
	}

	startedAt := time.Now()
	log.Printf("[sync] start survey sync: surveyPeriodID=%s, pageLength=%d, regionFullCode=%s", surveyPeriodID, maxSyncPageLength, requestedRegionFullCode)

	survey, err := s.surveyRepo.FindBySurveyPeriodID(surveyPeriodID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("survey period %s not found", surveyPeriodID)
		}
		return nil, err
	}

	creds := dto.FasihCredentials{
		Cookie:    survey.Cookie,
		XSRFToken: survey.XSRFToken,
	}

	effectiveLength := maxSyncPageLength

	regionFilterParam := dto.FasihAssignmentExtraParam{
		SurveyPeriodID: survey.SurveyPeriodID,
	}
	levelCount, regionID, err := s.resolveDatatableRegionFilter(ctx, survey, creds, requestedRegionFullCode)
	if err != nil {
		return nil, err
	}

	switch levelCount {
	case 1:
		regionFilterParam.Region1ID = &regionID
	case 2:
		regionFilterParam.Region2ID = &regionID
	case 3:
		regionFilterParam.Region3ID = &regionID
	case 4:
		regionFilterParam.Region4ID = &regionID
	case 5:
		regionFilterParam.Region5ID = &regionID
	case 6:
		regionFilterParam.Region6ID = &regionID
	default:
		return nil, apperrors.NewHttpError(http.StatusBadRequest, "level region survey tidak valid")
	}

	result := &dto.SyncSurveyAssignmentsResponse{}
	batch := 0
	skippedUnchanged := 0
	matchedAssignments := 0

	allowedStatuses := map[string]struct{}{
		strings.ToUpper("DRAFT"):                 {},
		strings.ToUpper("SUBMITTED BY Pencacah"): {},
		strings.ToUpper("COMPLETED BY Pengawas"): {},
		strings.ToUpper("APPROVED BY Pengawas"):  {},
		strings.ToUpper("REJECTED BY Pengawas"):  {},
		strings.ToUpper("REVOKED BY Pengawas"):   {},
	}
	syncStatuses := []string{"DRAFT", "SUBMITTED BY Pencacah", "COMPLETED BY Pengawas", "APPROVED BY Pengawas", "REJECTED BY Pengawas", "REVOKED BY Pengawas"}

	for _, syncStatus := range syncStatuses {
		start := 0
		for {
			batch++
			datatableReq := dto.FasihDatatableRequest{
				Start:  start,
				Length: effectiveLength,
				AssignmentExtraParam: dto.FasihAssignmentExtraParam{
					SurveyPeriodID:        regionFilterParam.SurveyPeriodID,
					AssignmentStatusAlias: syncStatus,
					Region1ID:             regionFilterParam.Region1ID,
					Region2ID:             regionFilterParam.Region2ID,
					Region3ID:             regionFilterParam.Region3ID,
					Region4ID:             regionFilterParam.Region4ID,
					Region5ID:             regionFilterParam.Region5ID,
					Region6ID:             regionFilterParam.Region6ID,
				},
			}

			datatableResp, err := s.fasihService.GetAssignmentDatatable(ctx, creds, datatableReq)
			if err != nil {
				return nil, err
			}

			result.TotalAssignments += datatableResp.TotalHit
			log.Printf("[sync] region=%s status=%s batch=%d fetched=%d offset=%d total=%d", requestedRegionFullCode, syncStatus, batch, len(datatableResp.SearchData), start, datatableResp.TotalHit)
			if len(datatableResp.SearchData) == 0 {
				log.Printf("[sync] region=%s status=%s no data returned on batch=%d, stop paging", requestedRegionFullCode, syncStatus, batch)
				break
			}

			for _, row := range datatableResp.SearchData {
				rowStatus := strings.ToUpper(strings.TrimSpace(row.AssignmentStatusAlias))
				if _, ok := allowedStatuses[rowStatus]; !ok {
					continue
				}

				regionFullCode, regionLevel1, regionLevel2, regionLevel3, regionLevel4, regionLevel5, regionLevel6 := extractRegionLevelCodes(row.Region)

				if requestedRegionFullCode != "" {
					rowRegionFullCode := ""
					if regionFullCode != nil {
						rowRegionFullCode = strings.TrimSpace(*regionFullCode)
					}

					if rowRegionFullCode != requestedRegionFullCode {
						continue
					}
				}

				matchedAssignments++

				existingAssignment, err := s.assignmentRepo.FindByAssignmentID(row.ID)
				hasExistingAssignment := err == nil
				if err != nil && err != gorm.ErrRecordNotFound {
					return nil, err
				}

				openedAt := sql.NullTime{}
				if hasExistingAssignment && existingAssignment.OpenedAt.Valid {
					openedAt = existingAssignment.OpenedAt
				}

				startedAt := sql.NullTime{}
				if hasExistingAssignment && existingAssignment.StartedAt.Valid {
					startedAt = existingAssignment.StartedAt
				}

				submittedAt, err := parseFlexibleTime(row.DateCreated)
				if err != nil {
					return nil, fmt.Errorf("parse submittedAt for assignment %s: %w", row.ID, err)
				}

				revisedAt, err := parseRevisedAt(row.DateModified, submittedAt)
				if err != nil {
					return nil, fmt.Errorf("parse revisedAt for assignment %s: %w", row.ID, err)
				}

				assignment := &models.Assignment{
					SurveyPeriodID: survey.SurveyPeriodID,
					AssignmentID:   row.ID,
					Status:         models.AssignmentStatusCodeFromAlias(row.AssignmentStatusAlias),
					RegionFullCode: regionFullCode,
					RegionLevel1:   regionLevel1,
					RegionLevel2:   regionLevel2,
					RegionLevel3:   regionLevel3,
					RegionLevel4:   regionLevel4,
					RegionLevel5:   regionLevel5,
					RegionLevel6:   regionLevel6,
					Latitude:       row.Latitude,
					Longitude:      row.Longitude,
					OpenedAt:       openedAt,
					StartedAt:      startedAt,
					SubmittedAt:    submittedAt,
					RevisedAt:      revisedAt,
				}
				if err := s.assignmentRepo.Upsert(assignment); err != nil {
					return nil, err
				}
				result.SavedAssignments++

				if hasExistingAssignment && existingAssignment.RevisedAt.Equal(revisedAt) && existingAssignment.StartedAt.Valid {
					skippedUnchanged++
					continue
				}

				historyResp, err := s.fasihService.GetAssignmentHistoryByID(ctx, creds, dto.FasihAssignmentByIDRequest{AssignmentID: row.ID})
				if err != nil {
					return nil, err
				}
				logs, err := extractLogsFromHistory(row.ID, historyResp.Data)
				if err != nil {
					return nil, err
				}
				if err := s.logRepo.ReplaceByAssignmentID(row.ID, logs); err != nil {
					return nil, err
				}
				result.SavedLogs += len(logs)

				assignmentResp, err := s.fasihService.GetAssignmentByID(ctx, creds, dto.FasihAssignmentByIDRequest{AssignmentID: row.ID})
				if err != nil {
					return nil, err
				}

				openedAtFromDetail, hasOpenedAtFromDetail, err := extractOpenedAtFromDetail(assignmentResp.Data.Data)
				if err != nil {
					return nil, err
				}
				startedAtFromDetail, hasStartedAtFromDetail, err := extractStartedAtFromDetail(assignmentResp.Data.Data)
				if err != nil {
					return nil, err
				}

				shouldUpsertAssignment := false
				if hasOpenedAtFromDetail {
					assignment.OpenedAt = sql.NullTime{Time: openedAtFromDetail, Valid: true}
					shouldUpsertAssignment = true
				}
				if hasStartedAtFromDetail {
					assignment.StartedAt = sql.NullTime{Time: startedAtFromDetail, Valid: true}
					shouldUpsertAssignment = true
				}
				if shouldUpsertAssignment {
					if err := s.assignmentRepo.Upsert(assignment); err != nil {
						return nil, err
					}
				}

				answers, err := extractAnswersFromDetail(row.ID, assignmentResp.Data.Data)
				if err != nil {
					return nil, err
				}
				if err := s.answerRepo.ReplaceByAssignmentID(row.ID, answers); err != nil {
					return nil, err
				}
				result.SavedAnswers += len(answers)

				if result.SavedAssignments%10 == 0 {
					log.Printf("[sync] progress assignments=%d/%d logs=%d answers=%d skipped=%d", result.SavedAssignments, result.TotalAssignments, result.SavedLogs, result.SavedAnswers, skippedUnchanged)
				}
			}

			start += len(datatableResp.SearchData)
			if start >= datatableResp.TotalHit {
				log.Printf("[sync] region=%s status=%s reached total assignments at offset=%d", requestedRegionFullCode, syncStatus, start)
				break
			}
		}
	}

	result.TotalAssignments = matchedAssignments

	if err := s.surveyRepo.UpdateSurveyRegionAssignmentCounts(surveyPeriodID); err != nil {
		return nil, err
	}

	log.Printf("[sync] completed in %s: total=%d savedAssignments=%d savedLogs=%d savedAnswers=%d skipped=%d", time.Since(startedAt).Round(time.Second), result.TotalAssignments, result.SavedAssignments, result.SavedLogs, result.SavedAnswers, skippedUnchanged)

	return result, nil
}

func (s *SurveyService) GetSyncRegionTargets(surveyPeriodID string) ([]string, error) {
	regions, err := s.surveyRepo.FindBySurveyPeriodIDWithFilter(surveyPeriodID, repositories.AssignmentRegionFilter{})
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, len(regions))
	targets := make([]string, 0, len(regions))
	for _, region := range regions {
		fullCode := strings.TrimSpace(region.FullCode)
		if fullCode == "" {
			continue
		}

		if _, ok := seen[fullCode]; ok {
			continue
		}

		seen[fullCode] = struct{}{}
		targets = append(targets, fullCode)
	}

	sort.Strings(targets)
	if len(targets) == 0 {
		return nil, apperrors.NewHttpError(http.StatusBadRequest, "data region survey belum tersedia, lakukan sync region terlebih dahulu")
	}

	return targets, nil
}

func (s *SurveyService) resolveDatatableRegionFilter(
	ctx context.Context,
	survey *models.Survey,
	creds dto.FasihCredentials,
	requestedRegionFullCode string,
) (int, string, error) {
	groupID := ""
	if survey.RegionGroupID != nil {
		groupID = strings.TrimSpace(*survey.RegionGroupID)
	}

	if groupID == "" {
		surveyResp, err := s.fasihService.GetSurveyByID(ctx, creds, dto.FasihSurveyByIDRequest{SurveyID: survey.SurveyID})
		if err != nil {
			return 0, "", err
		}

		groupID = strings.TrimSpace(surveyResp.Data.RegionGroupID)
		if groupID == "" {
			return 0, "", apperrors.NewHttpError(http.StatusBadRequest, "region group id survey tidak ditemukan")
		}
	}

	levelCount := 0
	if survey.RegionLevelCount != nil {
		levelCount = *survey.RegionLevelCount
	}

	if levelCount < 1 || levelCount > 6 {
		metadataResp, err := s.fasihService.GetRegionMetadata(ctx, creds, dto.FasihRegionMetadataRequest{GroupID: groupID})
		if err != nil {
			return 0, "", err
		}

		levelCount = metadataResp.Data.LevelCount
	}

	if levelCount < 1 || levelCount > 6 {
		return 0, "", apperrors.NewHttpError(http.StatusBadRequest, "level region survey tidak valid")
	}

	requested := strings.TrimSpace(requestedRegionFullCode)
	if requested == "" {
		return 0, "", apperrors.NewHttpError(http.StatusBadRequest, "region_full_code wajib diisi")
	}

	parentFullCode := ""
	if levelCount > 1 {
		regions, err := s.surveyRepo.FindBySurveyPeriodIDWithFilter(survey.SurveyPeriodID, repositories.AssignmentRegionFilter{
			RegionFullCode: requested,
		})
		if err != nil {
			return 0, "", err
		}

		if len(regions) == 0 {
			return 0, "", apperrors.NewHttpError(http.StatusBadRequest, "region_full_code tidak ditemukan di data region survey")
		}

		var parentErr error
		parentFullCode, parentErr = buildParentFullCodeForLevel(regions[0], levelCount)
		if parentErr != nil {
			return 0, "", parentErr
		}
	}

	listResp, err := s.fasihService.GetRegionsByLevel(ctx, creds, dto.FasihRegionListRequest{
		GroupID:        groupID,
		Level:          levelCount,
		ParentFullCode: parentFullCode,
	})
	if err != nil {
		return 0, "", err
	}

	for _, item := range listResp.Data {
		if strings.TrimSpace(item.FullCode) != requested {
			continue
		}

		regionID := strings.TrimSpace(item.ID)
		if regionID == "" {
			return 0, "", apperrors.NewHttpError(http.StatusBadRequest, "region id tidak ditemukan")
		}

		return levelCount, regionID, nil
	}

	return 0, "", apperrors.NewHttpError(http.StatusBadRequest, fmt.Sprintf("region_full_code harus region level %d", levelCount))
}

func buildParentFullCodeForLevel(region models.Region, levelCount int) (string, error) {
	if levelCount <= 1 {
		return "", nil
	}

	levels := []*string{region.Level1, region.Level2, region.Level3, region.Level4, region.Level5}
	var parts []string
	for index := 0; index < levelCount-1; index++ {
		if index >= len(levels) || levels[index] == nil {
			return "", apperrors.NewHttpError(http.StatusBadRequest, fmt.Sprintf("kode level %d region tidak tersedia", index+1))
		}

		trimmed := strings.TrimSpace(*levels[index])
		if trimmed == "" {
			return "", apperrors.NewHttpError(http.StatusBadRequest, fmt.Sprintf("kode level %d region kosong", index+1))
		}

		parts = append(parts, trimmed)
	}

	return strings.Join(parts, ""), nil
}

func parseRevisedAt(raw string, fallback time.Time) (time.Time, error) {
	if strings.TrimSpace(raw) == "" {
		return fallback, nil
	}

	parsed, err := parseFlexibleTime(raw)
	if err != nil {
		return time.Time{}, err
	}

	return parsed, nil
}

func extractRegionLevelCodes(region dto.FasihRegion) (*string, *string, *string, *string, *string, *string, *string) {
	fullCode := regionCodePtr(region.Level1.FullCode)
	level1 := regionCodePtr(region.Level1.Code)

	if region.Level1.Level2 == nil {
		return fullCode, level1, nil, nil, nil, nil, nil
	}

	fullCode = regionCodePtr(region.Level1.Level2.FullCode)
	level2 := regionCodePtr(region.Level1.Level2.Code)
	if region.Level1.Level2.Level3 == nil {
		return fullCode, level1, level2, nil, nil, nil, nil
	}

	fullCode = regionCodePtr(region.Level1.Level2.Level3.FullCode)
	level3 := regionCodePtr(region.Level1.Level2.Level3.Code)
	if region.Level1.Level2.Level3.Level4 == nil {
		return fullCode, level1, level2, level3, nil, nil, nil
	}

	fullCode = regionCodePtr(region.Level1.Level2.Level3.Level4.FullCode)
	level4 := regionCodePtr(region.Level1.Level2.Level3.Level4.Code)
	if region.Level1.Level2.Level3.Level4.Level5 == nil {
		return fullCode, level1, level2, level3, level4, nil, nil
	}

	fullCode = regionCodePtr(region.Level1.Level2.Level3.Level4.Level5.FullCode)
	level5 := regionCodePtr(region.Level1.Level2.Level3.Level4.Level5.Code)
	if region.Level1.Level2.Level3.Level4.Level5.Level6 == nil {
		return fullCode, level1, level2, level3, level4, level5, nil
	}

	fullCode = regionCodePtr(region.Level1.Level2.Level3.Level4.Level5.Level6.FullCode)
	level6 := regionCodePtr(region.Level1.Level2.Level3.Level4.Level5.Level6.Code)
	return fullCode, level1, level2, level3, level4, level5, level6
}

func regionCodePtr(code string) *string {
	trimmed := strings.TrimSpace(code)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func parseSurveyDatePtr(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}

	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func appendLevelCode(levels [6]*string, level int, code string) [6]*string {
	updated := levels
	if level < 1 || level > len(updated) {
		return updated
	}

	trimmed := strings.TrimSpace(code)
	if trimmed == "" {
		return updated
	}

	value := trimmed
	updated[level-1] = &value
	return updated
}

func appendLevelLabel(labels [6]*string, level int, name string) [6]*string {
	updated := labels
	if level < 1 || level > len(updated) {
		return updated
	}

	trimmed := strings.TrimSpace(name)
	if trimmed == "" || trimmed == "-" {
		return updated
	}

	value := trimmed
	updated[level-1] = &value
	return updated
}

func extractLevelLabels(levels []dto.FasihRegionMetadataLevel) [6]*string {
	var labels [6]*string

	for _, level := range levels {
		if level.ID < 1 || level.ID > len(labels) {
			continue
		}

		name := strings.TrimSpace(level.Name)
		if name == "" || name == "-" {
			continue
		}

		value := name
		labels[level.ID-1] = &value
	}

	return labels
}

func dedupeRegionNodes(nodes []regionNode) []regionNode {
	if len(nodes) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(nodes))
	result := make([]regionNode, 0, len(nodes))

	for _, node := range nodes {
		if _, ok := seen[node.fullCode]; ok {
			continue
		}
		seen[node.fullCode] = struct{}{}
		result = append(result, node)
	}

	return result
}

func shouldSkipRegionItem(item dto.FasihRegionItem) bool {
	name := strings.TrimSpace(item.Name)
	return name == "-"
}

func extractLogsFromHistory(assignmentID string, items []dto.FasihAssignmentHistory) ([]models.Log, error) {
	logs := make([]models.Log, 0)
	seen := make(map[string]struct{})
	for _, item := range items {
		if len(item.Paradata) == 0 {
			continue
		}

		var payload dto.FasihParadataPayload
		if err := json.Unmarshal(item.Paradata, &payload); err != nil {
			return nil, err
		}

		for _, entity := range payload.ActionLogEntities {
			t, err := parseFlexibleTime(entity.Timestamp)
			if err != nil {
				return nil, err
			}

			hash := models.BuildLogEventHash(assignmentID, entity.Action, entity.Latitude, entity.Longitude, t)
			if _, ok := seen[hash]; ok {
				continue
			}

			seen[hash] = struct{}{}

			logs = append(logs, models.Log{
				AssignmentID: assignmentID,
				EventHash:    hash,
				Action:       entity.Action,
				Latitude:     entity.Latitude,
				Longitude:    entity.Longitude,
				ActionedAt:   t,
			})
		}
	}

	return logs, nil
}

func extractAnswersFromDetail(assignmentID string, raw dto.FasihJSON) ([]models.Answer, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var payload dto.FasihAnswerPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}

	answers := make([]models.Answer, 0)
	for _, item := range payload.Answers {
		if item.DataKey == "" {
			continue
		}

		answeredAt, hasCreated := parseDynamicTime(item.CreatedAt)
		revisedAt, hasUpdated := parseDynamicTime(item.UpdatedAt)
		if !hasCreated && !hasUpdated {
			continue
		}
		if !hasCreated && hasUpdated {
			answeredAt = revisedAt
		}
		if hasCreated && !hasUpdated {
			revisedAt = answeredAt
		}

		answers = append(answers, models.Answer{
			AssignmentID: assignmentID,
			Name:         item.DataKey,
			AnsweredAt:   answeredAt,
			RevisedAt:    revisedAt,
		})
	}

	return answers, nil
}

func extractOpenedAtFromDetail(raw dto.FasihJSON) (time.Time, bool, error) {
	if len(raw) == 0 {
		return time.Time{}, false, nil
	}

	var payload dto.FasihAnswerPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return time.Time{}, false, err
	}

	openedAt, hasOpenedAt := parseDynamicTime(payload.CreatedAt)
	return openedAt, hasOpenedAt, nil
}

func extractStartedAtFromDetail(raw dto.FasihJSON) (time.Time, bool, error) {
	if len(raw) == 0 {
		return time.Time{}, false, nil
	}

	var payload dto.FasihAnswerPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return time.Time{}, false, err
	}

	var earliestStartAt time.Time
	hasStartAt := false
	for _, item := range payload.Answers {
		key := strings.ToLower(strings.TrimSpace(item.DataKey))
		if key == "" || !strings.Contains(key, "mulai") {
			continue
		}

		startAt, ok := parseDynamicTime(item.Answer)
		if !ok {
			continue
		}

		if !hasStartAt || startAt.Before(earliestStartAt) {
			earliestStartAt = startAt
			hasStartAt = true
		}
	}

	if hasStartAt {
		return earliestStartAt, true, nil
	}

	return time.Time{}, false, nil
}

func parseDynamicTime(v interface{}) (time.Time, bool) {
	switch t := v.(type) {
	case nil:
		return time.Time{}, false
	case float64:
		return time.UnixMilli(int64(t)), true
	case int64:
		return time.UnixMilli(t), true
	case int:
		return time.UnixMilli(int64(t)), true
	case json.Number:
		i, err := t.Int64()
		if err != nil {
			return time.Time{}, false
		}
		return time.UnixMilli(i), true
	case string:
		parsed, err := parseFlexibleTime(t)
		if err != nil {
			return time.Time{}, false
		}
		return parsed, true
	default:
		return time.Time{}, false
	}
}

func parseOptionalFlexibleTime(raw string) (time.Time, bool, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return time.Time{}, false, nil
	}

	parsed, err := parseFlexibleTime(trimmed)
	if err != nil {
		return time.Time{}, false, err
	}

	return parsed, true, nil
}

func trimmedPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func inferRegionLevelCount(regions []models.Region) int {
	maxLevel := 0
	for _, region := range regions {
		switch {
		case region.Level6 != nil && strings.TrimSpace(*region.Level6) != "":
			if maxLevel < 6 {
				maxLevel = 6
			}
		case region.Level5 != nil && strings.TrimSpace(*region.Level5) != "":
			if maxLevel < 5 {
				maxLevel = 5
			}
		case region.Level4 != nil && strings.TrimSpace(*region.Level4) != "":
			if maxLevel < 4 {
				maxLevel = 4
			}
		case region.Level3 != nil && strings.TrimSpace(*region.Level3) != "":
			if maxLevel < 3 {
				maxLevel = 3
			}
		case region.Level2 != nil && strings.TrimSpace(*region.Level2) != "":
			if maxLevel < 2 {
				maxLevel = 2
			}
		case region.Level1 != nil && strings.TrimSpace(*region.Level1) != "":
			if maxLevel < 1 {
				maxLevel = 1
			}
		}
	}

	return maxLevel
}

func parseFlexibleTime(raw string) (time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("empty time value")
	}

	if ms, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		return time.UnixMilli(ms), nil
	}

	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"Jan 2, 2006, 3:04:05 PM",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, trimmed); err == nil {
			return parsed, nil
		}
	}

	layoutsWithoutTimezone := []string{
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05.999999999",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.999999999",
	}
	for _, layout := range layoutsWithoutTimezone {
		if parsed, err := time.ParseInLocation(layout, trimmed, defaultGMTPlus8); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported time format: %s", raw)
}
