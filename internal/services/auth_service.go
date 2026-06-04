package services

import (
	"jejak/internal/dto"
	apperrors "jejak/internal/errors"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userService *UserService
	jwtService  *JWTService
	bpsService  *BPSService
}

func NewAuthService(userService *UserService, jwtService *JWTService, bpsService *BPSService) *AuthService {
	return &AuthService{
		userService: userService,
		jwtService:  jwtService,
		bpsService:  bpsService,
	}
}

func (s *AuthService) Login(payload *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userService.GetUserByEmailOrUsername(payload.Identifier)
	if err != nil {
		return nil, apperrors.ErrUserNotFound
	}

	if user.Password == nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	// compare hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(payload.Password)); err != nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Roles)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID, user.Roles)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Refresh(refreshToken string) (string, error) {
	_, claims, err := s.jwtService.ParseToken(refreshToken)
	if err != nil {
		return "", apperrors.ErrInvalidToken
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", apperrors.ErrInvalidToken
	}

	var roles []string
	if raw, ok := claims["roles"].([]interface{}); ok {
		for _, v := range raw {
			if r, ok := v.(string); ok {
				roles = append(roles, r)
			}
		}
	}

	return s.jwtService.GenerateAccessToken(userID, roles)
}

func (s *AuthService) LoginBPS(token string) (*dto.LoginResponse, error) {
	userInfo, err := s.bpsService.GetUserInfo(token)
	if err != nil {
		return nil, apperrors.ErrInvalidToken
	}

	user, err := s.userService.GetUserByEmail(userInfo.Email)
	if err != nil {
		return nil, apperrors.ErrUserNotFound
	}

	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Roles)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID, user.Roles)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
