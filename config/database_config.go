package config

import "fmt"

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	Driver   string
}

func LoadDatabaseConfig() (*DatabaseConfig, error) {
	cfg := &DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", ""),
		Password: getEnv("DB_PASSWORD", ""),
		Name:     getEnv("DB_NAME", ""),
		Driver:   getEnv("DB_DRIVER", "postgres"),
	}

	if cfg.User == "" {
		return nil, fmt.Errorf("DB_USER is required")
	}
	if cfg.Name == "" {
		return nil, fmt.Errorf("DB_NAME is required")
	}

	return cfg, nil
}
