package config

import "fmt"

type FasihConfig struct {
	BaseURL string
}

func LoadFasihConfig() (*FasihConfig, error) {
	cfg := &FasihConfig{
		BaseURL: getEnv("FASIH_BASE_URL", "https://fasih-sm.bps.go.id"),
	}

	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("FASIH_BASE_URL is required")
	}

	return cfg, nil
}
