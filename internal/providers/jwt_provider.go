package providers

import (
	"jejak/config"
	"jejak/internal/services"
	"time"
)

func NewJWTProvider(
	cfg *config.AuthConfig,
) *services.JWTService {
	return services.NewJWTService(
		cfg.JWTSecret,
		15*time.Minute,
		7*24*time.Hour,
	)
}
