package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"jejak/config"
	"jejak/internal/dto"
	apperrors "jejak/internal/errors"
	"jejak/internal/services"
	"net"
	"net/url"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"golang.org/x/net/publicsuffix"
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

func (h *AuthHandler) resolveCookieDomain() string {
	appURL := strings.TrimSpace(h.appConfig.URL)
	if appURL == "" {
		return ""
	}

	parsedURL, err := url.Parse(appURL)
	if err != nil || parsedURL.Hostname() == "" {
		parsedURL, err = url.Parse("https://" + appURL)
		if err != nil {
			return ""
		}
	}

	hostname := strings.TrimSpace(parsedURL.Hostname())
	if hostname == "" || strings.EqualFold(hostname, "localhost") || net.ParseIP(hostname) != nil {
		return ""
	}

	// Use EffectiveTLDPlusOne to correctly handle multi-level suffixes like .co.id, .go.id, .org.uk
	// e.g., "app.example.co.id" → "example.co.id", "api.example.com" → "example.com"
	cookieDomain, err := publicsuffix.EffectiveTLDPlusOne(hostname)
	if err != nil || cookieDomain == "" {
		// Fallback for local/internal domains: use two-segment domain
		parts := strings.Split(hostname, ".")
		if len(parts) < 2 {
			return ""
		}
		cookieDomain = strings.Join(parts[len(parts)-2:], ".")
	}

	if cookieDomain == "" {
		return ""
	}

	return "." + cookieDomain
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
		if cookieDomain := h.resolveCookieDomain(); cookieDomain != "" {
			cookie.Domain = cookieDomain
		}
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
		if cookieDomain := h.resolveCookieDomain(); cookieDomain != "" {
			cookie.Domain = cookieDomain
		}
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
