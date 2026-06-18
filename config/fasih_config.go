package config

import (
	"fmt"
	"strconv"
)

type FasihConfig struct {
	BaseURL              string
	HttpTimeoutSeconds   int
	HttpMaxRetries       int
	HttpRetryBaseDelayMs int
}

func LoadFasihConfig() (*FasihConfig, error) {
	cfg := &FasihConfig{
		BaseURL:              getEnv("FASIH_BASE_URL", "https://fasih-sm.bps.go.id"),
		HttpTimeoutSeconds:   getEnvIntWithFallback("FASIH_HTTP_TIMEOUT_SECONDS", "FASIH_DATATABLE_TIMEOUT_SECONDS", 75),
		HttpMaxRetries:       getEnvIntWithFallback("FASIH_HTTP_MAX_RETRIES", "FASIH_DATATABLE_MAX_RETRIES", 3),
		HttpRetryBaseDelayMs: getEnvIntWithFallback("FASIH_HTTP_RETRY_BASE_DELAY_MS", "FASIH_DATATABLE_RETRY_BASE_DELAY_MS", 1500),
	}

	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("FASIH_BASE_URL is required")
	}

	if cfg.HttpTimeoutSeconds < 1 {
		cfg.HttpTimeoutSeconds = 75
	}
	if cfg.HttpMaxRetries < 1 {
		cfg.HttpMaxRetries = 1
	}
	if cfg.HttpRetryBaseDelayMs < 100 {
		cfg.HttpRetryBaseDelayMs = 1500
	}

	return cfg, nil
}

func getEnvInt(key string, fallback int) int {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvIntWithFallback(primaryKey, fallbackKey string, fallback int) int {
	value := getEnv(primaryKey, "")
	if value == "" {
		value = getEnv(fallbackKey, "")
	}

	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
