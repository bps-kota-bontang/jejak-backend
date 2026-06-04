package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"jejak/internal/dto"
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
	assignments, err := s.assignmentRepo.FindBySurveyPeriodID(surveyPeriodID)
	if err != nil {
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
			return nil, ctx.Err()
		default:
		}

		analysis, err := s.assignmentSvc.AnalyzeAssignment(ctx, assignment.AssignmentID)
		if err != nil {
			return nil, fmt.Errorf("analyze assignment %s: %w", assignment.AssignmentID, err)
		}

		result.AnalyzedAssignments++
		result.Assignments = append(result.Assignments, *analysis)
	}

	return result, nil
}

func (s *SurveyService) CreateSurvey(req dto.CreateSurveyRequest) error {
	survey := &models.Survey{
		SurveyID:       req.SurveyID,
		SurveyPeriodID: req.SurveyPeriodID,
		XSRFToken:      req.XSRFToken,
		Cookie:         req.Cookie,
	}

	return s.surveyRepo.Upsert(survey)
}

func (s *SurveyService) SyncSurveyAssignments(ctx context.Context, req dto.SyncSurveyAssignmentsRequest) (*dto.SyncSurveyAssignmentsResponse, error) {
	startedAt := time.Now()
	log.Printf("[sync] start survey sync: surveyPeriodID=%s, pageLength=%d", req.SurveyPeriodID, maxSyncPageLength)

	survey, err := s.surveyRepo.FindBySurveyPeriodID(req.SurveyPeriodID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("survey period %s not found", req.SurveyPeriodID)
		}
		return nil, err
	}

	creds := dto.FasihCredentials{
		Cookie:    survey.Cookie,
		XSRFToken: survey.XSRFToken,
	}

	effectiveLength := maxSyncPageLength

	result := &dto.SyncSurveyAssignmentsResponse{}
	start := 0
	batch := 0
	skippedUnchanged := 0

	for {
		batch++
		datatableReq := dto.FasihDatatableRequest{
			Start:  start,
			Length: effectiveLength,
			AssignmentExtraParam: dto.FasihAssignmentExtraParam{
				SurveyPeriodID:            survey.SurveyPeriodID,
				AssignmentErrorStatusType: req.AssignmentErrorStatusType,
				AssignmentStatusAlias:     req.AssignmentStatusAlias,
				FilterTargetType:          req.FilterTargetType,
			},
		}

		datatableResp, err := s.fasihService.GetAssignmentDatatable(ctx, creds, datatableReq)
		if err != nil {
			return nil, err
		}

		result.TotalAssignments = datatableResp.TotalHit
		log.Printf("[sync] batch=%d fetched=%d offset=%d total=%d", batch, len(datatableResp.SearchData), start, datatableResp.TotalHit)
		if len(datatableResp.SearchData) == 0 {
			log.Printf("[sync] no data returned on batch=%d, stop paging", batch)
			break
		}

		for _, row := range datatableResp.SearchData {
			existingAssignment, err := s.assignmentRepo.FindByAssignmentID(row.ID)
			hasExistingAssignment := err == nil
			if err != nil && err != gorm.ErrRecordNotFound {
				return nil, err
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
				Latitude:       row.Latitude,
				Longitude:      row.Longitude,
				OpenedAt:       sql.NullTime{},
				SubmittedAt:    submittedAt,
				RevisedAt:      revisedAt,
			}
			if err := s.assignmentRepo.Upsert(assignment); err != nil {
				return nil, err
			}
			result.SavedAssignments++

			if hasExistingAssignment && existingAssignment.RevisedAt.Equal(revisedAt) {
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
			log.Printf("[sync] reached total assignments at offset=%d", start)
			break
		}
	}

	log.Printf("[sync] completed in %s: total=%d savedAssignments=%d savedLogs=%d savedAnswers=%d skipped=%d", time.Since(startedAt).Round(time.Second), result.TotalAssignments, result.SavedAssignments, result.SavedLogs, result.SavedAnswers, skippedUnchanged)

	return result, nil
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

func extractLogsFromHistory(assignmentID string, items []dto.FasihAssignmentHistory) ([]models.Log, error) {
	logs := make([]models.Log, 0)
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

			logs = append(logs, models.Log{
				AssignmentID: assignmentID,
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

	return time.Time{}, fmt.Errorf("unsupported time format: %s", raw)
}
