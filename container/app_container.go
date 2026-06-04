package container

import (
	"jejak/config"

	"github.com/gofiber/fiber/v3"
)

type AppContainer struct {
	App    *fiber.App
	Config *config.AppConfig
}
