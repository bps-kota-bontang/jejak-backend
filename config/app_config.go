package config

import (
	"log"
	"os"
	"regexp"

	"github.com/joho/godotenv"
)

var hashPattern = regexp.MustCompile(`^[0-9a-fA-F]{8,}$`)

func shortenBuildHash(build string) string {
	if len(build) > 7 && hashPattern.MatchString(build) {
		return build[:7]
	}

	return build
}

type AppConfig struct {
	Name  string
	Env   string
	Port  string
	URL   string
	Build string
}

func LoadAppConfig() (*AppConfig, error) {
	// Cek APP_ENV sebelum load .env agar tidak tertimpa
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	// Load .env hanya di luar production
	if env != "production" {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: .env file not found: %v", err)
		} else {
			log.Println("Loaded .env file")
		}
	}

	return &AppConfig{
		Name:  getEnv("APP_NAME", "jejak"),
		Env:   getEnv("APP_ENV", env),
		Port:  getEnv("APP_PORT", "3000"),
		URL:   getEnv("APP_URL", "http://localhost:3000"),
		Build: shortenBuildHash(getEnv("APP_BUILD", "development")),
	}, nil
}

// IsProduction returns true when running in production environment.
func (c *AppConfig) IsProduction() bool {
	return c.Env == "production"
}
