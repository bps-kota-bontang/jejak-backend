//go:build wireinject
// +build wireinject

package main

import (
	"jejak/container"
	"jejak/internal/di"

	"github.com/google/wire"
)

// InitializeScheduler builds SchedulerContainer
func InitializeScheduler() (*container.SchedulerContainer, error) {
	wire.Build(di.SchedulerSet)
	return &container.SchedulerContainer{}, nil
}
