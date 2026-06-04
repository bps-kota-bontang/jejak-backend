//go:build wireinject
// +build wireinject

package main

import (
	"jejak/container"
	"jejak/internal/di"

	"github.com/google/wire"
)

func InitializeSurvey() (*container.SurveyContainer, error) {
	wire.Build(di.SurveySet)
	return &container.SurveyContainer{}, nil
}
