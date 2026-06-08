package routes

import (
	"jejak/internal/handlers"

	"github.com/gofiber/fiber/v3"
)

func RegisterAssignmentRoutes(router fiber.Router, handler *handlers.AssignmentHandler) {
	assignments := router.Group("/assignments")
	assignments.Get("/:assignmentId", handler.GetByAssignmentID)
	assignments.Get("/:assignmentId/logs", handler.GetLogsByAssignmentID)
	assignments.Post("/:assignmentId/analyze", handler.AnalyzeAssignment)
}
