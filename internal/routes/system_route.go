package routes

import (
	"jejak/internal/handlers"

	"github.com/gofiber/fiber/v3"
)

func RegisterSystemRoutes(router fiber.Router, handler *handlers.SystemHandler) {
	system := router.Group("/system")
	system.Get("/features", handler.GetFeatures)
	system.Get("/features/fasih-authorization", handler.GetFasihAuthorization)
}
