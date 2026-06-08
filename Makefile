.PHONY: wire run run-worker run-scheduler run-survey run-survey-sync-region build build-worker build-scheduler dev dev-worker dev-scheduler generate hot

# Generate wire_gen.go
wire:
	go run github.com/google/wire/cmd/wire gen ./cmd/app ./cmd/worker ./cmd/scheduler ./cmd/survey

# Go generate (buat jalankan semua go:generate, termasuk wire)
generate:
	go generate ./...

# Run app
run:
	go run ./cmd/app

# Seed data from RKK.xlsx
run-seed:
	go run ./cmd/seed

# Seed realization from FA Detail.xlsx
run-fa-detail:
	go run ./cmd/fa-detail

# Run worker
run-worker:
	go run ./cmd/worker

# Run scheduler
run-scheduler:
	go run ./cmd/scheduler

# Run survey command
run-survey:
	go run ./cmd/survey $(ARGS)

# Run survey region sync command
run-survey-sync-region:
	go run ./cmd/survey sync-region

# Build app binary
build:
	go build -o myapp ./cmd/app

# Build worker binary
build-worker:
	go build -o myworker ./cmd/worker

# Build scheduler binary
build-scheduler:
	go build -o myscheduler ./cmd/scheduler

# Dev mode: auto generate wire, then run app
dev:
	make generate
	make run

# Dev mode for worker
dev-worker:
	make generate
	make run-worker

# Dev mode for scheduler
dev-scheduler:
	make generate
	make run-scheduler

# Hot reload mode with air
hot:
	air