package services

import (
	"context"
	"fmt"
	"sort"
	"time"

	"jejak/internal/dto"
	"jejak/internal/helpers"
	"jejak/internal/models"
	"jejak/internal/repositories"
)

type AssignmentService struct {
	assignmentRepo repositories.AssignmentRepository
	logRepo        repositories.LogRepository
	answerRepo     repositories.AnswerRepository
	locationRepo   repositories.LocationRepository
}

func NewAssignmentService(
	assignmentRepo repositories.AssignmentRepository,
	logRepo repositories.LogRepository,
	answerRepo repositories.AnswerRepository,
	locationRepo repositories.LocationRepository,
) *AssignmentService {
	return &AssignmentService{
		assignmentRepo: assignmentRepo,
		logRepo:        logRepo,
		answerRepo:     answerRepo,
		locationRepo:   locationRepo,
	}
}

// AnalyzeAssignment analyzes one assignment, persists location proportions,
// and maps each answer row to the inferred location_id.
func (s *AssignmentService) AnalyzeAssignment(ctx context.Context, assignmentID string) (*dto.AssignmentSurveyAnalysis, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	assignment, err := s.assignmentRepo.FindByAssignmentID(assignmentID)
	if err != nil {
		return nil, err
	}

	analysis, err := s.analyzeAssignmentCore(*assignment)
	if err != nil {
		return nil, fmt.Errorf("analyze assignment %s: %w", assignmentID, err)
	}

	return &analysis, nil
}

func (s *AssignmentService) analyzeAssignmentCore(assignment models.Assignment) (dto.AssignmentSurveyAnalysis, error) {
	analysis := dto.AssignmentSurveyAnalysis{
		AssignmentID:   assignment.AssignmentID,
		SurveyPeriodID: assignment.SurveyPeriodID,
		Locations:      make([]dto.LocationAnswerStat, 0),
	}

	logs, err := s.logRepo.FindByAssignmentID(assignment.AssignmentID)
	if err != nil {
		return analysis, err
	}
	if len(logs) == 0 {
		if err := s.locationRepo.ReplaceByAssignmentID(assignment.AssignmentID, nil); err != nil {
			return analysis, err
		}
		return analysis, nil
	}

	answers, err := s.answerRepo.FindByAssignmentID(assignment.AssignmentID)
	if err != nil {
		return analysis, err
	}
	if len(answers) == 0 {
		if err := s.locationRepo.ReplaceByAssignmentID(assignment.AssignmentID, nil); err != nil {
			return analysis, err
		}
		return analysis, nil
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].ActionedAt.Before(logs[j].ActionedAt)
	})

	locationCounts := map[string]*dto.LocationAnswerStat{}
	answerCanonicalID := map[string]string{}

	for _, answer := range answers {
		answerTime := getAnswerEventTime(answer)
		closestLog := nearestLogAtOrBefore(logs, answerTime)

		lat := helpers.RoundCoordinate(closestLog.Latitude)
		lon := helpers.RoundCoordinate(closestLog.Longitude)
		canonicalID := helpers.CanonicalIDFromCoordinates(lat, lon)

		stat, ok := locationCounts[canonicalID]
		if !ok {
			stat = &dto.LocationAnswerStat{
				CanonicalID: canonicalID,
				Latitude:    lat,
				Longitude:   lon,
			}
			locationCounts[canonicalID] = stat
		}

		stat.AnswerCount++
		analysis.TotalAnswers++
		answerCanonicalID[answer.ID] = canonicalID
	}

	for _, stat := range locationCounts {
		if analysis.TotalAnswers > 0 {
			stat.Proportion = float64(stat.AnswerCount) / float64(analysis.TotalAnswers)
		}
		analysis.Locations = append(analysis.Locations, *stat)
	}

	sort.Slice(analysis.Locations, func(i, j int) bool {
		return analysis.Locations[i].AnswerCount > analysis.Locations[j].AnswerCount
	})

	locations := make([]models.Location, 0, len(analysis.Locations))
	for _, stat := range analysis.Locations {
		locations = append(locations, models.Location{
			AssignmentID: analysis.AssignmentID,
			CanonicalID:  stat.CanonicalID,
			Latitude:     stat.Latitude,
			Longitude:    stat.Longitude,
			AnswerCount:  stat.AnswerCount,
			Proportion:   stat.Proportion,
		})
	}

	if err := s.locationRepo.ReplaceByAssignmentID(assignment.AssignmentID, locations); err != nil {
		return analysis, err
	}

	persistedLocations, err := s.locationRepo.FindByAssignmentID(assignment.AssignmentID)
	if err != nil {
		return analysis, err
	}

	locationIDByCanonicalID := map[string]string{}
	for _, item := range persistedLocations {
		locationIDByCanonicalID[item.CanonicalID] = item.ID
	}

	for _, answer := range answers {
		canonicalID, ok := answerCanonicalID[answer.ID]
		if !ok {
			continue
		}
		locationID, ok := locationIDByCanonicalID[canonicalID]
		if !ok {
			continue
		}
		locationIDCopy := locationID
		if err := s.answerRepo.UpdateLocationID(answer.ID, &locationIDCopy); err != nil {
			return analysis, err
		}
	}

	return analysis, nil
}

func getAnswerEventTime(answer models.Answer) time.Time {
	if !answer.AnsweredAt.IsZero() {
		return answer.AnsweredAt
	}
	if !answer.RevisedAt.IsZero() {
		return answer.RevisedAt
	}
	return answer.CreatedAt
}

func nearestLogAtOrBefore(logs []models.Log, t time.Time) models.Log {
	if len(logs) == 0 {
		return models.Log{}
	}

	candidate := logs[0]
	for _, lg := range logs {
		if !lg.ActionedAt.After(t) {
			candidate = lg
			continue
		}
		break
	}

	return candidate
}
