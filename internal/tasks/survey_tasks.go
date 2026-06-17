package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"jejak/internal/dto"
	"jejak/internal/services"

	"github.com/hibiken/asynq"
)

// syncTaskTimeout is the maximum duration a sync task is allowed to run.
// Asynq will consider the task timed out and retry if it exceeds this limit.
const syncTaskTimeout = 60 * time.Minute

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
		log.Printf("[task][error] failed to marshal payload type=%s surveyPeriodID=%s regionFullCode=%s err=%v", taskType, payload.SurveyPeriodID, payload.RegionFullCode, err)
		return nil, fmt.Errorf("marshal survey task payload: %w", err)
	}

	// Build a deterministic task ID to prevent duplicate tasks for the same
	// survey+region combination from being queued simultaneously.
	taskID := taskType + ":" + payload.SurveyPeriodID
	if strings.TrimSpace(payload.RegionFullCode) != "" {
		taskID += ":" + strings.TrimSpace(payload.RegionFullCode)
	}

	task := asynq.NewTask(taskType, data,
		asynq.TaskID(taskID),
		asynq.Timeout(syncTaskTimeout),
		asynq.MaxRetry(0),
		asynq.Unique(syncTaskTimeout),
	)
	info, err := client.EnqueueContext(ctx, task)
	if err != nil {
		// If a task with the same ID is already pending or active, treat it as
		// a no-op rather than returning an error to the caller.
		if errors.Is(err, asynq.ErrTaskIDConflict) {
			log.Printf("[task] skipped duplicate enqueue for taskID=%s", taskID)
			return nil, nil
		}
		log.Printf("[task][error] enqueue failed type=%s taskID=%s surveyPeriodID=%s regionFullCode=%s err=%v", taskType, taskID, payload.SurveyPeriodID, payload.RegionFullCode, err)
		return nil, fmt.Errorf("enqueue %s task: %w", taskType, err)
	}

	log.Printf("[task] enqueued taskID=%s queue=%s", taskID, info.Queue)
	return info, nil
}

func handleSurveySyncTask(surveyService *services.SurveyService, regionScoped bool) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) error {
		var payload SurveyTaskPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			log.Printf("[worker][sync][error] failed to unmarshal payload err=%v", err)
			return fmt.Errorf("unmarshal survey sync payload: %w", err)
		}

		log.Printf("[worker][sync] start surveyPeriodID=%s regionFullCode=%s", payload.SurveyPeriodID, payload.RegionFullCode)
		_, err := surveyService.SyncSurveyAssignments(ctx, payload.SurveyPeriodID, dto.SyncSurveyAssignmentsRequest{
			RegionFullCode: payload.RegionFullCode,
		})
		if err != nil {
			log.Printf("[worker][sync][error] sync failed surveyPeriodID=%s regionFullCode=%s err=%v", payload.SurveyPeriodID, payload.RegionFullCode, err)
			return nil
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
			log.Printf("[worker][analyze][error] failed to unmarshal payload err=%v", err)
			return nil
		}

		log.Printf("[worker][analyze] start surveyPeriodID=%s regionFullCode=%s", payload.SurveyPeriodID, payload.RegionFullCode)
		if regionScoped {
			_, err := surveyService.AnalyzeSurveyByRegion(ctx, payload.SurveyPeriodID, payload.RegionFullCode)
			if err != nil {
				log.Printf("[worker][analyze][error] analyze by region failed surveyPeriodID=%s regionFullCode=%s err=%v", payload.SurveyPeriodID, payload.RegionFullCode, err)
				return nil
			}
			log.Printf("[worker][analyze] completed region task surveyPeriodID=%s regionFullCode=%s", payload.SurveyPeriodID, payload.RegionFullCode)
			return nil
		}

		_, err := surveyService.AnalyzeSurvey(ctx, payload.SurveyPeriodID)
		if err != nil {
			log.Printf("[worker][analyze][error] analyze failed surveyPeriodID=%s err=%v", payload.SurveyPeriodID, err)
			return nil
		}
		log.Printf("[worker][analyze] completed survey task surveyPeriodID=%s", payload.SurveyPeriodID)
		return nil
	}
}
