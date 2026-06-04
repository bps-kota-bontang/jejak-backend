package config

import "fmt"

type AuthConfig struct {
	GateURL   string
	JWTSecret string
	GateID    string
}

func LoadAuthConfig() (*AuthConfig, error) {
	cfg := &AuthConfig{
		GateURL:   getEnv("AUTH_GATE_URL", ""),
		JWTSecret: getEnv("AUTH_JWT_SECRET", ""),
		GateID:    getEnv("AUTH_GATE_ID", ""),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("AUTH_JWT_SECRET is required")
	}

	return cfg, nil
}
