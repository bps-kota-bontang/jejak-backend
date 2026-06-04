package di

import (
	"jejak/internal/services"

	"github.com/google/wire"
)

var ServiceSet = wire.NewSet(
	services.NewUserService,
	services.NewAuthService,
	services.NewBPSService,
	services.NewFasihService,
	services.NewSurveyService,
	services.NewAssignmentService,
)
