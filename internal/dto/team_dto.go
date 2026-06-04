package dto

type TeamResponse struct {
	ID      uint     `json:"id"`
	Name    string   `json:"name"`
	UserIDs []string `json:"user_ids"`
}

type CreateTeamRequest struct {
	Name    string   `json:"name" validate:"required"`
	UserIDs []string `json:"user_ids"`
}

type UpdateTeamRequest struct {
	Name    *string  `json:"name,omitempty"`
	UserIDs []string `json:"user_ids,omitempty"`
}
