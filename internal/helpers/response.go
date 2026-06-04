package helpers

// Response is the standard API response envelope used across all endpoints.
type Response struct {
	Data    any      `json:"data"`
	Message string   `json:"message"`
	Errors  []string `json:"errors,omitempty"`
	Meta    any      `json:"meta,omitempty"`
}
