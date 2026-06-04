package services

import (
	"time"

	apperrors "jejak/internal/errors"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type AccessClaims struct {
	UserID string   `json:"user_id"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

func NewJWTService(secret string, accessTTL, refreshTTL time.Duration) *JWTService {
	return &JWTService{
		Secret:          secret,
		AccessTokenTTL:  accessTTL,
		RefreshTokenTTL: refreshTTL,
	}
}

func (j *JWTService) GenerateAccessToken(userID string, roles []string) (string, error) {
	claims := AccessClaims{
		UserID: userID,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.Secret))
}

func (j *JWTService) GenerateRefreshToken(userID string, roles []string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"roles": roles,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(j.RefreshTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.Secret))
}

// ParseToken parses any token and returns MapClaims (used for refresh flow).
func (j *JWTService) ParseToken(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperrors.ErrInvalidToken
		}
		return []byte(j.Secret), nil
	})
	if err != nil || !token.Valid {
		return nil, nil, apperrors.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, apperrors.ErrInvalidToken
	}

	return token, claims, nil
}

// ValidateAccessToken validates an access token and returns the userID and roles.
func (j *JWTService) ValidateAccessToken(tokenString string) (string, []string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperrors.ErrInvalidToken
		}
		return []byte(j.Secret), nil
	})
	if err != nil || !token.Valid {
		return "", nil, apperrors.ErrInvalidToken
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok {
		return "", nil, apperrors.ErrInvalidToken
	}

	return claims.UserID, claims.Roles, nil
}
