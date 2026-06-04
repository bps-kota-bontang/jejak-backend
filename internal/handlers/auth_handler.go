package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"jejak/config"
	"jejak/internal/dto"
	apperrors "jejak/internal/errors"
	"jejak/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	appConfig  *config.AppConfig
	authConfig *config.AuthConfig
	service    *services.AuthService
	validate   *validator.Validate
}

func NewAuthHandler(appConfig *config.AppConfig, authConfig *config.AuthConfig, service *services.AuthService, validate *validator.Validate) *AuthHandler {
	return &AuthHandler{
		appConfig:  appConfig,
		authConfig: authConfig,
		service:    service,
		validate:   validate,
	}
}

// setRefreshTokenCookie adalah helper function untuk set cookie refresh token
func (h *AuthHandler) setRefreshTokenCookie(c fiber.Ctx, value string, maxAge int) {
	isProd := h.appConfig.IsProduction()

	cookie := &fiber.Cookie{
		Name:     "refresh_token",
		Value:    value,
		Path:     "/",
		HTTPOnly: true,
		Secure:   isProd,
		SameSite: "None",
		MaxAge:   maxAge,
	}

	if isProd {
		cookie.Domain = ".bpsbontang.com"
	}

	c.Cookie(cookie)
}

// setStateCookie adalah helper function untuk set cookie state (untuk SSO)
func (h *AuthHandler) setStateCookie(c fiber.Ctx, value string, maxAge int) {
	isProd := h.appConfig.IsProduction()

	cookie := &fiber.Cookie{
		Name:     "state",
		Value:    value,
		Path:     "/",
		HTTPOnly: true,
		Secure:   isProd,
		SameSite: "None",
		MaxAge:   maxAge,
	}

	if isProd {
		cookie.Domain = ".bpsbontang.com"
	}

	c.Cookie(cookie)
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	return respondError(c, apperrors.NewHttpError(fiber.StatusGone, "Password login is disabled. Use SSO login instead"))
}

func (h *AuthHandler) RefreshToken(c fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")

	if refreshToken == "" {
		return respondError(c, apperrors.ErrMissingToken)
	}

	newAccess, err := h.service.Refresh(refreshToken)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, fiber.Map{"access_token": newAccess}, "Access token refreshed successfully")
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	// Expire refresh token cookie
	h.setRefreshTokenCookie(c, "", -1)

	return respondOK(c, nil, "Logout successful")
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (h *AuthHandler) RedirectSSO(c fiber.Ctx) error {
	state, err := generateState()
	if err != nil {
		return respondError(c, apperrors.NewHttpError(fiber.StatusInternalServerError, "Failed to generate state"))
	}

	// Set state cookie (10 menit untuk CSRF protection)
	h.setStateCookie(c, state, 10*60)

	redirectURL := fmt.Sprintf(
		"%s/api/v1/auth/sso?state=%s&service_id=%s",
		h.authConfig.GateURL,
		state,
		h.authConfig.GateID,
	)

	return c.Redirect().To(redirectURL)
}

func (h *AuthHandler) LoginSSO(c fiber.Ctx) error {
	cookieState := c.Cookies("state")
	var payload dto.LoginSSORequest
	if err := c.Bind().Body(&payload); err != nil {
		return respondValidation(c, err)
	}

	if err := h.validate.Struct(&payload); err != nil {
		return respondValidation(c, err)
	}

	if payload.State != cookieState {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "Invalid or expired state"))
	}

	tokens, err := h.service.LoginBPS(payload.Token)
	if err != nil {
		return respondError(c, err)
	}

	// Set refresh token cookie (7 hari)
	h.setRefreshTokenCookie(c, tokens.RefreshToken, 7*24*60*60)

	return respondOK(c, fiber.Map{"access_token": tokens.AccessToken}, "Login successful")
}
