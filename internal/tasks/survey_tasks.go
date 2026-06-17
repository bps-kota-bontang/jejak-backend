package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"jejak/internal/dto"
	"jejak/internal/services"

	"github.com/hibiken/asynq"
)

const (
	TypeSurveySync          = "survey:sync"
	TypeSurveySyncRegion    = "survey:sync_region"
	TypeSurveyAnalyze       = "survey:analyze"
	TypeSurveyAnalyzeRegion = "survey:analyze_region"
)

type SurveyTaskPayload struct {
	SurveyPeriodID string `json:"survey_period_id"`
	RegionFullCode string `json:"region_full_code,omitempty"`
}

func EnqueueSurveySync(ctx context.Context, client *asynq.Client, surveyPeriodID string, regionFullCode string) (*asynq.TaskInfo, error) {
	payload := SurveyTaskPayload{
		SurveyPeriodID: surveyPeriodID,
		RegionFullCode: regionFullCode,
	}
	return enqueueSurveyTask(ctx, client, TypeSurveySync, payload)
}

func EnqueueSurveySyncRegion(ctx context.Context, client *asynq.Client, surveyPeriodID string, regionFullCode string) (*asynq.TaskInfo, error) {
	payload := SurveyTaskPayload{
		SurveyPeriodID: surveyPeriodID,
		RegionFullCode: regionFullCode,
	}
	return enqueueSurveyTask(ctx, client, TypeSurveySyncRegion, payload)
}

func EnqueueSurveyAnalyze(ctx context.Context, client *asynq.Client, surveyPeriodID string) (*asynq.TaskInfo, error) {
	payload := SurveyTaskPayload{SurveyPeriodID: surveyPeriodID}
	return enqueueSurveyTask(ctx, client, TypeSurveyAnalyze, payload)
}

func EnqueueSurveyAnalyzeRegion(ctx context.Context, client *asynq.Client, surveyPeriodID string, regionFullCode string) (*asynq.TaskInfo, error) {
	payload := SurveyTaskPayload{
		SurveyPeriodID: surveyPeriodID,
		RegionFullCode: regionFullCode,
	}
	return enqueueSurveyTask(ctx, client, TypeSurveyAnalyzeRegion, payload)
}

func RegisterSurveyTaskHandlers(mux *asynq.ServeMux, surveyService *services.SurveyService) {
	mux.HandleFunc(TypeSurveySync, handleSurveySyncTask(surveyService, false))
	mux.HandleFunc(TypeSurveySyncRegion, handleSurveySyncTask(surveyService, true))
	mux.HandleFunc(TypeSurveyAnalyze, handleSurveyAnalyzeTask(surveyService, false))
	mux.HandleFunc(TypeSurveyAnalyzeRegion, handleSurveyAnalyzeTask(surveyService, true))
}

func enqueueSurveyTask(ctx context.Context, client *asynq.Client, taskType string, payload SurveyTaskPayload) (*asynq.TaskInfo, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal survey task payload: %w", err)
	}

	task := asynq.NewTask(taskType, data)
	info, err := client.EnqueueContext(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("enqueue %s task: %w", taskType, err)
	}

	return info, nil
}

func handleSurveySyncTask(surveyService *services.SurveyService, regionScoped bool) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) error {
		var payload SurveyTaskPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("unmarshal survey sync payload: %w", err)
		}

		log.Printf("[worker][sync] start surveyPeriodID=%s regionFullCode=%s", payload.SurveyPeriodID, payload.RegionFullCode)
		_, err := surveyService.SyncSurveyAssignments(ctx, payload.SurveyPeriodID, dto.SyncSurveyAssignmentsRequest{
			RegionFullCode: payload.RegionFullCode,
		})
		if err != nil {
			return fmt.Errorf("sync surveyPeriodID=%s regionFullCode=%s: %w", payload.SurveyPeriodID, payload.RegionFullCode, err)
		}
		if regionScoped {
			log.Printf("[worker][sync] completed region task surveyPeriodID=%s regionFullCode=%s", payload.SurveyPeriodID, payload.RegionFullCode)
			return nil
		}
		log.Printf("[worker][sync] completed survey task surveyPeriodID=%s", payload.SurveyPeriodID)
		return nil
	}
}

func handleSurveyAnalyzeTask(surveyService *services.SurveyService, regionScoped bool) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) error {
		var payload SurveyTaskPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("unmarshal survey analyze payload: %w", err)
		}

		log.Printf("[worker][analyze] start surveyPeriodID=%s regionFullCode=%s", payload.SurveyPeriodID, payload.RegionFullCode)
		if regionScoped {
			_, err := surveyService.AnalyzeSurveyByRegion(ctx, payload.SurveyPeriodID, payload.RegionFullCode)
			if err != nil {
				return fmt.Errorf("analyze surveyPeriodID=%s regionFullCode=%s: %w", payload.SurveyPeriodID, payload.RegionFullCode, err)
			}
			log.Printf("[worker][analyze] completed region task surveyPeriodID=%s regionFullCode=%s", payload.SurveyPeriodID, payload.RegionFullCode)
			return nil
		}

		_, err := surveyService.AnalyzeSurvey(ctx, payload.SurveyPeriodID)
		if err != nil {
			return fmt.Errorf("analyze surveyPeriodID=%s: %w", payload.SurveyPeriodID, err)
		}
		log.Printf("[worker][analyze] completed survey task surveyPeriodID=%s", payload.SurveyPeriodID)
		return nil
	}
}
