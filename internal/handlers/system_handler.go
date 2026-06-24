package handlers

import (
	"strings"

	"jejak/internal/dto"
	apperrors "jejak/internal/errors"
	"jejak/internal/repositories"
	"jejak/internal/services"

	"github.com/gofiber/fiber/v3"
)

type SystemHandler struct {
	fasihService *services.FasihService
	surveyRepo   repositories.SurveyRepository
}

func NewSystemHandler(fasihService *services.FasihService, surveyRepo repositories.SurveyRepository) *SystemHandler {
	return &SystemHandler{fasihService: fasihService, surveyRepo: surveyRepo}
}

func (h *SystemHandler) GetFeatures(c fiber.Ctx) error {
	userAgent := strings.TrimSpace(c.Get("User-Agent"))
	response := dto.SystemFeaturesResponse{FasihAvailable: h.fasihService.IsAvailableWithUserAgent(c.Context(), userAgent)}
	return respondOK(c, response, "System features retrieved successfully")
}

func (h *SystemHandler) GetFasihAuthorization(c fiber.Ctx) error {
	surveyPeriodID := strings.TrimSpace(c.Query("survey_period_id"))
	if surveyPeriodID == "" {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "survey_period_id wajib diisi"))
	}

	survey, err := h.surveyRepo.FindBySurveyPeriodID(surveyPeriodID)
	if err != nil {
		return respondError(c, apperrors.NewHttpError(fiber.StatusNotFound, "Survey tidak ditemukan untuk survey_period_id tersebut"))
	}

	response := dto.SystemFasihAuthorizationResponse{
		SurveyPeriodID: surveyPeriodID,
		FasihAuthorized: h.fasihService.IsAuthorizedForSurveyPeriodWithUserAgent(
			c.Context(),
			dto.FasihCredentials{
				Cookie:    survey.Cookie,
				XSRFToken: survey.XSRFToken,
			},
			surveyPeriodID,
			strings.TrimSpace(c.Get("User-Agent")),
		),
	}

	return respondOK(c, response, "System features retrieved successfully")
}
