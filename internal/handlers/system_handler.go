package handlers

import (
	"jejak/internal/dto"
	"jejak/internal/services"

	"github.com/gofiber/fiber/v3"
)

type SystemHandler struct {
	fasihService *services.FasihService
}

func NewSystemHandler(fasihService *services.FasihService) *SystemHandler {
	return &SystemHandler{fasihService: fasihService}
}

func (h *SystemHandler) GetFeatures(c fiber.Ctx) error {
	return respondOK(c, dto.SystemFeaturesResponse{FasihAvailable: h.fasihService.IsAvailable(c.Context())}, "System features retrieved successfully")
}
