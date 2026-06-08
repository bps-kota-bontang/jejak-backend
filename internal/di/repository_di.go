package di

import (
	"jejak/internal/repositories"

	"github.com/google/wire"
)

var RepositorySet = wire.NewSet(
	repositories.NewUserRepository,
	repositories.NewSurveyRepository,
	repositories.NewAssignmentRepository,
	repositories.NewLogRepository,
	repositories.NewAnswerRepository,
	repositories.NewLocationRepository,
	repositories.NewAreaRepository,
)
