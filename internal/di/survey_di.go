package di

import (
	"jejak/container"

	"github.com/google/wire"
)

var SurveySet = wire.NewSet(
	ConfigSet,
	ProviderSet,
	RepositorySet,
	ServiceSet,
	container.NewSurveyContainer,
)
