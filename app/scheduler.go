package app

import (
	"fmt"
	"jejak/config"
	"jejak/container"
	"log"

	"github.com/hibiken/asynq"
)

// NewAsynqScheduler creates a new asynq scheduler
func NewAsynqScheduler(
	appConfig *config.AppConfig,
	schedulerConfig *config.SchedulerConfig,
	redisConfig *config.RedisConfig,
	asynqClient *asynq.Client,
) (*container.SchedulerContainer, error) {
	redisClientOpt := asynq.RedisClientOpt{
		Addr: redisConfig.Host + ":" + redisConfig.Port,
	}

	defer asynqClient.Close()

	scheduler := asynq.NewScheduler(redisClientOpt, nil)

	log.Println("Task registered successfully")

	if err := scheduler.Start(); err != nil {
		return nil, fmt.Errorf("failed to start scheduler: %w", err)
	}

	return &container.SchedulerContainer{
		Scheduler: scheduler,
	}, nil
}
