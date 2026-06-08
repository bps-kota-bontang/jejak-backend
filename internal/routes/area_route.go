package routes

import (
	"jejak/internal/handlers"

	"github.com/gofiber/fiber/v3"
)

func RegisterAreaRoutes(router fiber.Router, handler *handlers.AreaHandler) {
	areas := router.Group("/areas")
	areas.Get("/", handler.GetAll)
	areas.Post("/upload", handler.UploadGeoJSON)
	areas.Post("/", handler.Create)
	areas.Get("/:id", handler.GetByID)
	areas.Put("/:id", handler.Update)
	areas.Delete("/:id", handler.Delete)
}
