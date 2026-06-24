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
	requestUserAgent := strings.TrimSpace(c.Get("User-Agent"))
	effectiveUserAgent, userAgentSource := h.fasihService.ResolveUserAgentForLog(requestUserAgent)
	log.Printf("[handler][system][features] check fasih availability requestUserAgent=%q effectiveUserAgentSource=%s effectiveUserAgent=%q", requestUserAgent, userAgentSource, effectiveUserAgent)
	response := dto.SystemFeaturesResponse{FasihAvailable: h.fasihService.IsAvailableWithUserAgent(c.Context(), requestUserAgent)}
	log.Printf("[handler][system][features] result fasihAvailable=%t requestUserAgent=%q effectiveUserAgentSource=%s effectiveUserAgent=%q", response.FasihAvailable, requestUserAgent, userAgentSource, effectiveUserAgent)
	return respondOK(c, response, "System features retrieved successfully")
}

func (h *SystemHandler) GetFasihAuthorization(c fiber.Ctx) error {
	surveyPeriodID := strings.TrimSpace(c.Query("survey_period_id"))
	if surveyPeriodID == "" {
		return respondError(c, apperrors.NewHttpError(fiber.StatusBadRequest, "survey_period_id wajib diisi"))
	}

	requestUserAgent := strings.TrimSpace(c.Get("User-Agent"))
	effectiveUserAgent, userAgentSource := h.fasihService.ResolveUserAgentForLog(requestUserAgent)
	log.Printf("[handler][system][fasih-authorization] check surveyPeriodID=%s requestUserAgent=%q effectiveUserAgentSource=%s effectiveUserAgent=%q", surveyPeriodID, requestUserAgent, userAgentSource, effectiveUserAgent)

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
	log.Printf("[handler][system][fasih-authorization] result surveyPeriodID=%s fasihAuthorized=%t requestUserAgent=%q effectiveUserAgentSource=%s effectiveUserAgent=%q", surveyPeriodID, response.FasihAuthorized, requestUserAgent, userAgentSource, effectiveUserAgent)

	return respondOK(c, response, "System features retrieved successfully")
}
