.PHONY: dev build up down test test-backend test-frontend integration-test clean

# Development
dev:
	@echo "Starting backend..."
	cd backend && go run ./cmd/server/ &
	@echo "Starting frontend..."
	cd frontend && npm run dev

# Docker
build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down

# Testing
test: test-backend test-frontend

test-backend:
	cd backend && go test ./... -v -race

test-frontend:
	cd frontend && npm test -- --run

# Integration test: analyze the backend project itself via the API
integration-test:
	@echo "Running integration test..."
	@./scripts/integration-test.sh

clean:
	cd backend && make clean
	docker compose down --rmi local -v 2>/dev/null || true
