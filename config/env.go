package config

import "os"

// getEnv returns the value of the environment variable named by key,
// or fallback if the variable is unset or empty.
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
