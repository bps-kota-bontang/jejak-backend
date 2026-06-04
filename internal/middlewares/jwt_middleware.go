package middlewares

import (
	"jejak/internal/helpers"
	"jejak/internal/services"
	"jejak/utils"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type JWTMiddleware struct {
	jwtService *services.JWTService
}

func NewJWTMiddleware(jwtService *services.JWTService) *JWTMiddleware {
	return &JWTMiddleware{jwtService}
}

func (m *JWTMiddleware) Protected() fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(helpers.Response{
				Data:    nil,
				Message: "Missing or malformed authorization header",
			})
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		userID, roles, err := m.jwtService.ValidateAccessToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(helpers.Response{
				Data:    nil,
				Message: "Invalid or expired token",
			})
		}

		c.Locals("user_id", userID)
		c.Locals("roles", roles)
		return c.Next()
	}
}

func (m *JWTMiddleware) RequireRoles(requiredRoles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		roles, _ := c.Locals("roles").([]string)

		if !utils.HasAnyRole(roles, requiredRoles...) {
			return c.Status(fiber.StatusForbidden).JSON(helpers.Response{
				Data:    nil,
				Message: "You are not authorized to access this resource",
			})
		}

		return c.Next()
	}
}
