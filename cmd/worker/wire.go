//go:build wireinject
// +build wireinject

package main

import (
	"jejak/container"
	"jejak/internal/di"

	"github.com/google/wire"
)

// InitializeWorker builds WorkerContainer
func InitializeWorker() (*container.WorkerContainer, error) {
	wire.Build(di.WorkerSet)
	return &container.WorkerContainer{}, nil
}
