package di

import (
	"jejak/internal/providers"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	providers.NewAsyncClient,
	providers.NewValidator,
	providers.NewRedisClient,
	providers.NewDBConnection,
	providers.NewJWTProvider,
)
