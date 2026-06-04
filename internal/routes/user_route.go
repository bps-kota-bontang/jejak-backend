package routes

import (
	"jejak/internal/handlers"

	"github.com/gofiber/fiber/v3"
)

// RegisterUserRoutes registers all user-related routes
func RegisterUserRoutes(router fiber.Router, handler *handlers.UserHandler) {
	user := router.Group("/users")
	user.Get("/me", handler.Me)
	user.Patch("/me/password", handler.UpdateMyPassword)
	user.Get("/", handler.GetAllUsers)
	user.Get("/roles", handler.GetRoleOptions)
	user.Post("/", handler.CreateUser)
	user.Get("/:id", handler.GetUserByID)
	user.Put("/:id", handler.UpdateUser)
	user.Delete("/:id", handler.DeleteUser)
}
