# Makefile for managing the Docker-based Go Nihongo API application
# This file provides targets to build, run, and manage the application using Docker Compose.

.PHONY: build-dev build-prod up-dev down-dev up-prod down-prod

# Build the Docker images for development
build-dev:
	docker-compose -f docker-compose.dev.yml build

# Build the Docker images for production
build-prod:
	docker-compose -f docker-compose.prod.yml build

# Development environment targets
# For development, we use docker-compose.dev.yml which may include volume mounts for hot-reload
up-dev:
	docker-compose -f docker-compose.dev.yml up -d

down-dev:
	docker-compose -f docker-compose.dev.yml down

# Production environment targets
# For production, we use docker-compose.prod.yml with optimized settings
up-prod:
	docker-compose -f docker-compose.prod.yml up -d

down-prod:
	docker-compose -f docker-compose.prod.yml down