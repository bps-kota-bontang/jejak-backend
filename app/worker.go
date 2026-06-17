package app

import (
	"jejak/config"
	"jejak/container"
	"jejak/internal/services"
	"jejak/internal/tasks"

	"github.com/hibiken/asynq"
)

// NewAsynqWorker initializes the Asynq worker with all necessary routes
func NewAsynqWorker(
	appConfig *config.AppConfig,
	redisConfig *config.RedisConfig,
	surveyService *services.SurveyService,
) (*container.WorkerContainer, error) {
	redisClientOpt := asynq.RedisClientOpt{
		Addr: redisConfig.Host + ":" + redisConfig.Port,
	}

	srv := asynq.NewServer(redisClientOpt, asynq.Config{
		Concurrency: 20,
		Queues: map[string]int{
			"critical": 12,
			"default":  5,
			"low":      3,
		},
	})

	mux := asynq.NewServeMux()
	tasks.RegisterSurveyTaskHandlers(mux, surveyService)

	return &container.WorkerContainer{
		Server: srv,
		Mux:    mux,
	}, nil
}
