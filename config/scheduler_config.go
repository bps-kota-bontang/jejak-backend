package config

type SchedulerConfig struct{}

func LoadSchedulerConfig() (*SchedulerConfig, error) {
	return &SchedulerConfig{}, nil
}
