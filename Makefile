.PHONY: help build run test lint fmt docs clean install check-all format-all lint-file

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

install: ## Install dependencies
	go mod download
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/DarthSim/hivemind@latest

build: ## Build the server binary
	go build -o bin/server ./cmd/server

run: ## Run the server
	go run ./cmd/server/main.go

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests and show coverage report
	go tool cover -html=coverage.out

lint: ## Run linter
	golangci-lint run ./...

fmt: ## Format code
	gofmt -w -s .
	go mod tidy

docs: ## Generate OpenAPI documentation
	swag init -g cmd/server/main.go -o ./docs

clean: ## Clean build artifacts
	rm -rf bin/ docs/ coverage.out todos.db

dev: docs run ## Generate docs and run server

dev-server: docs ## Start both backend and frontend dev servers with logging
	@mkdir -p .logs
	hivemind Procfile 2>&1 | tee dev-server.log

dev-logs: ## Tail the dev server logs
	tail -f dev-server.log

check-all: ## Run all linters, formatters, and type checks (backend + frontend)
	@echo "Running backend checks..." && $(MAKE) lint & \
	$(MAKE) fmt & \
	echo "Running frontend checks..." && cd frontend && npm run check & \
	wait

format-all: ## Format code for both backend and frontend
	@echo "Formatting backend..."
	@$(MAKE) fmt
	@echo "Formatting frontend..."
	@cd frontend && npm run format

lint-file: ## Lint and format a single file (requires FILE=path/to/file)
	@if [ -z "$(FILE)" ]; then \
		echo "Error: FILE variable is required. Usage: make lint-file FILE=path/to/file"; \
		exit 1; \
	fi; \
	if echo "$(FILE)" | grep -q '\.go$$'; then \
		echo "Go linting results for $(FILE):"; \
		gofmt -w -s "$(FILE)" 2>&1; \
		golangci-lint run "$(FILE)" 2>&1 || true; \
	elif echo "$(FILE)" | grep -qE '\.(ts|tsx)$$'; then \
		echo "TypeScript linting results for $(FILE):"; \
		if echo "$(FILE)" | grep -q '^frontend/'; then \
			FILE_REL=$${FILE#frontend/}; \
		else \
			FILE_REL="$(FILE)"; \
		fi; \
		cd frontend && npx prettier --write "$$FILE_REL" 2>&1 > /dev/null && npx tsgo --jsx react-jsx --noEmit "$$FILE_REL" || true; \
	else \
		echo "Unsupported file type: $(FILE)"; \
	fi

.DEFAULT_GOAL := help
