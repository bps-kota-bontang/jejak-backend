package di

import (
	"jejak/internal/middlewares"

	"github.com/google/wire"
)

var MiddlewareSet = wire.NewSet(
	middlewares.NewJWTMiddleware,
)
