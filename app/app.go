package app

import (
	"jejak/config"
	"jejak/container"
	"jejak/internal/handlers"
	"jejak/internal/middlewares"
	"jejak/internal/routes"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/static"
)

func NewFiberApp(
	AppConfig *config.AppConfig,
	JWTMiddleware *middlewares.JWTMiddleware,
	AuthHandler *handlers.AuthHandler,
	UserHandler *handlers.UserHandler,
	SurveyHandler *handlers.SurveyHandler,
	AssignmentHandler *handlers.AssignmentHandler,
	SystemHandler *handlers.SystemHandler,
	AreaHandler *handlers.AreaHandler,
) (*container.AppContainer, error) {
	App := fiber.New(
		fiber.Config{AppName: AppConfig.Name},
	)

	corsConfig := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Refresh-Attempt"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Content-Disposition"},
	}

	if AppConfig.IsProduction() {
		corsConfig.AllowOrigins = []string{AppConfig.URL}
	} else {
		corsConfig.AllowOrigins = []string{"http://localhost:5173"}
	}

	App.Use(cors.New(corsConfig))

	App.Get("/", func(c fiber.Ctx) error {
		return c.SendString("API Jejak (Build: " + AppConfig.Build + ")")
	})

	App.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Serve static assets such as GeoJSON boundary files with access-token protection.
	App.Use("/static", JWTMiddleware.Protected(), static.New("./public"))

	api := App.Group("/api")
	apiV1 := api.Group("/v1")

	// Public
	routes.RegisterAuthRoutes(apiV1, AuthHandler, JWTMiddleware)
	routes.RegisterSystemRoutes(apiV1, SystemHandler)

	// Protected
	protected := apiV1.Group("/", JWTMiddleware.Protected())
	routes.RegisterUserRoutes(protected, UserHandler)
	routes.RegisterSurveyRoutes(protected, SurveyHandler)
	routes.RegisterAssignmentRoutes(protected, AssignmentHandler)
	routes.RegisterAreaRoutes(protected, AreaHandler)

	return &container.AppContainer{
		App:    App,
		Config: AppConfig,
	}, nil
}
