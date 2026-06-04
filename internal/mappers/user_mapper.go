package mappers

import (
	"jejak/internal/dto"
	"jejak/internal/models"
)

func ToUserResponse(user *models.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Roles:     user.Roles,
		CreatedAt: user.CreatedAt,
	}
}

func ToUserModel(input *dto.CreateUserRequest) *models.User {
	return &models.User{
		Username: input.Username,
		Email:    input.Email,
		Roles:    input.Roles,
	}
}

func ApplyUserUpdateFromRequest(user *models.User, req *dto.UpdateUserRequest) {
	if req.Username != nil {
		user.Username = *req.Username
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Roles != nil {
		user.Roles = req.Roles
	}
}
