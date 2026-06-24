package handlers

import (
	"log"
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
	log.Printf("[handler][system][features] check fasih availability userAgent=%q", userAgent)
	response := dto.SystemFeaturesResponse{FasihAvailable: h.fasihService.IsAvailableWithUserAgent(c.Context(), userAgent)}
	log.Printf("[handler][system][features] result fasihAvailable=%t userAgent=%q", response.FasihAvailable, userAgent)
	return respondOK(c, response, "System features retrieved successfully")
}

func (h *SystemHandler) GetFasihAuthorization(c fiber.Ctx) error {
	surveyPeriodID := strings.TrimSpace(c.Query("survey_period_id"))
	if surveyPeriodID == "" {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "survey_period_id wajib diisi"))
	}

	requestUserAgent := strings.TrimSpace(c.Get("User-Agent"))
	log.Printf("[handler][system][fasih-authorization] check surveyPeriodID=%s userAgent=%q", surveyPeriodID, requestUserAgent)

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
			requestUserAgent,
		),
	}
	log.Printf("[handler][system][fasih-authorization] result surveyPeriodID=%s fasihAuthorized=%t userAgent=%q", surveyPeriodID, response.FasihAuthorized, requestUserAgent)

	return respondOK(c, response, "System features retrieved successfully")
}
