package errors

import "net/http"

// Sentinel errors for common business logic failures.
// Services should return these (or NewHttpError) instead of generic fmt.Errorf.
var (
	ErrUserNotFound       = NewHttpError(http.StatusNotFound, "User not found")
	ErrInvalidCredentials = NewHttpError(http.StatusUnauthorized, "Invalid credentials")
	ErrInvalidToken       = NewHttpError(http.StatusUnauthorized, "Invalid or expired token")
	ErrMissingToken       = NewHttpError(http.StatusUnauthorized, "Missing token")
)
