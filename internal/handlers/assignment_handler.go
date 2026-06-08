package handlers

import (
	apperrors "jejak/internal/errors"
	"jejak/internal/mappers"
	"jejak/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type AssignmentHandler struct {
	service  *services.AssignmentService
	validate *validator.Validate
}

func NewAssignmentHandler(service *services.AssignmentService, validate *validator.Validate) *AssignmentHandler {
	return &AssignmentHandler{
		service:  service,
		validate: validate,
	}
}

func (h *AssignmentHandler) GetByAssignmentID(c fiber.Ctx) error {
	assignmentID := c.Params("assignmentId")
	if assignmentID == "" {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "assignmentId is required"))
	}

	assignment, err := h.service.GetByAssignmentID(assignmentID)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, mappers.ToAssignmentResponse(assignment), "Assignment retrieved successfully")
}

func (h *AssignmentHandler) AnalyzeAssignment(c fiber.Ctx) error {
	assignmentID := c.Params("assignmentId")
	if assignmentID == "" {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "assignmentId is required"))
	}

	result, err := h.service.AnalyzeAssignment(c.Context(), assignmentID)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, result, "Assignment analyzed successfully")
}

func (h *AssignmentHandler) GetLogsByAssignmentID(c fiber.Ctx) error {
	assignmentID := c.Params("assignmentId")
	if assignmentID == "" {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "assignmentId is required"))
	}

	logs, err := h.service.GetLogsByAssignmentID(assignmentID)
	if err != nil {
		return respondError(c, err)
	}

	return respondOK(c, mappers.ToAssignmentLogPointResponses(logs), "Assignment logs retrieved successfully")
}
