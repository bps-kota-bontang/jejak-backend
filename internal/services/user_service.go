package services

import (
	"errors"
	"net/http"
	"slices"
	"strings"

	"jejak/internal/dto"
	apperrors "jejak/internal/errors"
	"jejak/internal/mappers"
	"jejak/internal/models"
	"jejak/internal/repositories"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo repositories.UserRepository
}

func NewUserService(
	userRepo repositories.UserRepository,
) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	return s.userRepo.FindByEmail(email)
}

func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	return s.userRepo.FindByUsername(username)
}

func (s *UserService) GetUserByEmailOrUsername(identifier string) (*models.User, error) {
	return s.userRepo.FindByEmailOrUsername(identifier)
}

func (s *UserService) GetUserByID(id string) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, err
	}
	return mappers.ToUserResponse(user), nil
}

func (s *UserService) GetAllUsers() ([]dto.UserResponse, error) {
	users, err := s.userRepo.FindAll()
	if err != nil {
		return nil, err
	}

	responses := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		resp := mappers.ToUserResponse(&user)
		if resp != nil {
			responses = append(responses, *resp)
		}
	}

	return responses, nil
}

func (s *UserService) GetRoleOptions() ([]string, error) {
	roles, err := s.userRepo.ListDistinctRoles()
	if err != nil {
		return nil, err
	}

	defaults := []string{"admin", "viewer", "operator", "ppk", "ketua"}
	for _, role := range defaults {
		if !slices.Contains(roles, role) {
			roles = append(roles, role)
		}
	}

	slices.Sort(roles)
	return roles, nil
}

func (s *UserService) GetAllPaginated(
	search string,
	page, perPage int,
	sortBy, sortOrder string,
	filters map[string][]string,
) ([]dto.UserResponse, int64, error) {

	if perPage > 100 {
		perPage = 100
	}

	var total int64
	if err := s.userRepo.Count(search, filters, &total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	users, err := s.userRepo.FindPaginated(search, perPage, offset, sortBy, sortOrder, filters)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		resp := mappers.ToUserResponse(&user)
		if resp != nil {
			responses = append(responses, *resp)
		}
	}

	return responses, total, nil
}

func (s *UserService) CreateUser(req *dto.CreateUserRequest) error {
	var userExisting *models.User

	userExisting, _ = s.userRepo.FindByEmail(req.Email)
	if userExisting != nil {
		return apperrors.NewHttpError(http.StatusConflict, "email is already taken")
	}

	userExisting, _ = s.userRepo.FindByUsername(req.Username)
	if userExisting != nil {
		return apperrors.NewHttpError(http.StatusConflict, "username is already taken")
	}

	user := mappers.ToUserModel(req)

	if req.Password != nil && strings.TrimSpace(*req.Password) != "" {
		passwordHashed, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		passwordHashedStr := string(passwordHashed)
		user.Password = &passwordHashedStr
	}

	if err := s.userRepo.Create(user); err != nil {
		return err
	}

	return nil
}

func (s *UserService) UpdateUser(id string, req *dto.UpdateUserRequest) error {
	if req.Email != nil {
		userExisting, _ := s.userRepo.FindByEmail(*req.Email)
		if userExisting != nil && userExisting.ID != id {
			return apperrors.NewHttpError(http.StatusConflict, "email is already in use by another user")
		}
	}

	if req.Username != nil {
		userExisting, _ := s.userRepo.FindByUsername(*req.Username)
		if userExisting != nil && userExisting.ID != id {
			return apperrors.NewHttpError(http.StatusConflict, "username is already in use by another user")
		}
	}

	user, err := s.userRepo.FindByIDIncludePassword(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrUserNotFound
		}
		return err
	}

	mappers.ApplyUserUpdateFromRequest(user, req)

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	return nil
}

func (s *UserService) UpdateMyPassword(userID string, req *dto.UpdateMyPasswordRequest) error {
	user, err := s.userRepo.FindByIDIncludePassword(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrUserNotFound
		}
		return err
	}

	if req.NewPassword == req.OldPassword {
		return apperrors.NewHttpError(http.StatusBadRequest, "new password must be different from old password")
	}

	if user.Password != nil {
		if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(req.OldPassword)); err != nil {
			return apperrors.NewHttpError(http.StatusBadRequest, "old password is incorrect")
		}
	}

	passwordHashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	passwordHashedStr := string(passwordHashed)
	user.Password = &passwordHashedStr

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	return nil
}

func (s *UserService) DeleteUser(id string) error {
	_, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrUserNotFound
		}
		return err
	}
	return s.userRepo.Delete(id)
}
