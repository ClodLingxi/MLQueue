.PHONY: help install run build test clean docker-up docker-down

help: ## Display this help message
	@echo "MLQueue - Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install: ## Install backend and frontend dependencies
	@echo "Installing backend dependencies..."
	go mod download
	@echo "Installing frontend dependencies..."
	cd web && npm install

run-backend: ## Run backend server
	@echo "Starting backend server..."
	go run main.go

run-frontend: ## Run frontend development server
	@echo "Starting frontend development server..."
	cd web && npm run dev

run: ## Run both backend and frontend (requires tmux)
	@echo "Starting services..."
	tmux new-session -d -s mlqueue 'make run-backend'
	tmux split-window -h 'cd web && npm run dev'
	tmux attach -t mlqueue

build: ## Build backend binary
	@echo "Building backend..."
	go build -o bin/mlqueue main.go

build-frontend: ## Build frontend for production
	@echo "Building frontend..."
	cd web && npm run build

test: ## Run backend tests
	@echo "Running tests..."
	go test -v ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf web/dist/

docker-up: ## Start services with Docker Compose
	@echo "Starting Docker services..."
	docker-compose up -d

docker-down: ## Stop Docker services
	@echo "Stopping Docker services..."
	docker-compose down

docker-logs: ## Show Docker logs
	docker-compose logs -f

db-init: ## Initialize database with sample data
	@echo "Initializing database..."
	psql -U mlqueue -d mlqueue_db -f scripts/init_db.sql

dev: ## Start development environment
	@echo "Starting development environment..."
	@echo "Make sure PostgreSQL and Redis are running!"
	@echo ""
	@echo "Backend will run on http://localhost:8080"
	@echo "Frontend will run on http://localhost:5173"
	@echo ""
	@$(MAKE) -j2 run-backend run-frontend
