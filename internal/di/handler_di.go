package di

import (
	"jejak/internal/handlers"

	"github.com/google/wire"
)

var HandlerSet = wire.NewSet(
	handlers.NewAuthHandler,
	handlers.NewUserHandler,
)
