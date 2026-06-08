package routes

import (
	"jejak/internal/handlers"

	"github.com/gofiber/fiber/v3"
)

func RegisterSurveyRoutes(router fiber.Router, handler *handlers.SurveyHandler) {
	surveys := router.Group("/surveys")
	surveys.Get("/", handler.GetAll)
	surveys.Post("/", handler.CreateSurvey)
	surveys.Put("/:surveyPeriodId", handler.UpdateSurvey)
	surveys.Get("/:surveyPeriodId", handler.GetBySurveyPeriodID)
	surveys.Get("/:surveyPeriodId/assignments", handler.GetAssignments)
	surveys.Get("/:surveyPeriodId/logs", handler.GetLogs)
	surveys.Get("/:surveyPeriodId/regions/metadata", handler.GetRegionMetadata)
	surveys.Get("/:surveyPeriodId/regions", handler.GetRegions)
	surveys.Post("/:surveyPeriodId/regions/sync", handler.SyncSurveyRegions)
	surveys.Post("/:surveyPeriodId/regions/:regionFullCode/assignments/sync", handler.SyncSurveyAssignmentsByRegion)
	surveys.Post("/:surveyPeriodId/regions/import", handler.ImportSurveyRegions)
	surveys.Post("/:surveyPeriodId/sync", handler.SyncSurveyAssignments)
	surveys.Post("/:surveyPeriodId/assignments/import", handler.ImportSurveyAssignments)
	surveys.Post("/:surveyPeriodId/regions/:regionFullCode/analyze", handler.AnalyzeSurveyByRegion)
	surveys.Post("/:surveyPeriodId/analyze", handler.AnalyzeSurvey)
}
