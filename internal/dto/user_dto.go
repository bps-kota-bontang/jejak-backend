package dto

import "time"

type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Roles     []string  `json:"roles"`
	TeamIDs   []uint    `json:"team_ids"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateUserRequest struct {
	Username string   `json:"username" validate:"required"`
	Email    string   `json:"email" validate:"required,email"`
	Password *string  `json:"password,omitempty" validate:"omitempty,min=8"`
	Roles    []string `json:"roles"`
}

type UpdateUserRequest struct {
	Username *string  `json:"username,omitempty"`
	Email    *string  `json:"email,omitempty" validate:"omitempty,email"`
	Roles    []string `json:"roles,omitempty"`
}

type UpdateMyPasswordRequest struct {
	OldPassword     string `json:"old_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}
