package app

import (
	"jejak/config"
	"jejak/container"
	"jejak/internal/handlers"
	"jejak/internal/middlewares"
	"jejak/internal/routes"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func NewFiberApp(
	AppConfig *config.AppConfig,
	JWTMiddleware *middlewares.JWTMiddleware,
	AuthHandler *handlers.AuthHandler,
	UserHandler *handlers.UserHandler,
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

	api := App.Group("/api")
	apiV1 := api.Group("/v1")

	// Public
	routes.RegisterAuthRoutes(apiV1, AuthHandler, JWTMiddleware)

	// Protected
	protected := apiV1.Group("/", JWTMiddleware.Protected())
	routes.RegisterUserRoutes(protected, UserHandler)

	return &container.AppContainer{
		App:    App,
		Config: AppConfig,
	}, nil
}
