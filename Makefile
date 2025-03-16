.PHONY: all build clean deps dev docker docker-build docker-push generate help lint mock run test vet proto

# Application name
APP_NAME := go-user-api
# Main package path
MAIN_PACKAGE := ./cmd
# Git commit hash
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
# Current version
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.1.0")
# Build date
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOVET := $(GOCMD) vet
GOMOD := $(GOCMD) mod
GOLINT := golangci-lint
GOCOVER := $(GOCMD) tool cover
GOMOCK := mockgen
GOTESTSUM := gotestsum

# Docker parameters
DOCKER_REGISTRY := 
DOCKER_IMAGE := $(APP_NAME)
DOCKER_TAG := $(VERSION)

# Proto parameters
PROTOC := protoc
PROTO_DIR := ./internal/interfaces/grpc/proto

# Environment variables
export GO111MODULE := on
export CGO_ENABLED := 0
export GOOS := linux

# Build flags
BUILD_FLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.commitHash=$(COMMIT_HASH) -X main.buildDate=$(BUILD_DATE)"

all: lint test build

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

build: ## Build the application
	$(GOBUILD) $(BUILD_FLAGS) -o $(APP_NAME) $(MAIN_PACKAGE)

clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -f $(APP_NAME)
	rm -rf ./bin

deps: ## Install dependencies
	$(GOMOD) download
	$(GOMOD) tidy

generate: ## Run go generate
	$(GOCMD) generate ./...

lint: ## Run linter
	$(GOLINT) run

test: ## Run tests
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCOVER) -html=coverage.out -o coverage.html

mock: ## Generate mocks
	@echo "Generating mocks..."
	$(GOMOCK) -source=./internal/domain/repository/user_repository.go -destination=./internal/domain/mocks/user_repository_mock.go -package=mocks UserRepository
	$(GOMOCK) -source=./internal/domain/usecase/user_usecase.go -destination=./internal/domain/mocks/user_usecase_mock.go -package=mocks UserUseCase

proto: ## Generate protobuf files
	@echo "Generating protobuf files..."
	$(PROTOC) --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/user.proto

vet: ## Run go vet
	$(GOVET) ./...

run: ## Run the application
	$(GOCMD) run $(MAIN_PACKAGE)

docker-build: ## Build docker image
	docker build -t $(DOCKER_REGISTRY)$(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-push: ## Push docker image
	docker push $(DOCKER_REGISTRY)$(DOCKER_IMAGE):$(DOCKER_TAG)

docker: docker-build docker-push ## Build and push docker image

dev: ## Run the application in development mode with hot reload
	air -c .air.toml

#db-create: ## Create database
#	@echo "Creating database..."
#	psql -h localhost -U postgres -c "CREATE DATABASE user_service"
#
#db-migrate: ## Run database migrations
#	@echo "Running migrations..."
#	migrate -path ./db/migrations -database "postgres://postgres:postgres@localhost:5432/user_service?sslmode=disable" up
#
#db-rollback: ## Rollback database migrations
#	@echo "Rolling back migrations..."
#	migrate -path ./db/migrations -database "postgres://postgres:postgres@localhost:5432/user_service?sslmode=disable" down 1
#
#db-reset: ## Reset database
#	@echo "Resetting database..."
#	migrate -path ./db/migrations -database "postgres://postgres:postgres@localhost:5432/user_service?sslmode=disable" drop
#	$(MAKE) db-migrate

docker-up: ## Start docker containers
	docker-compose up -d

docker-down: ## Stop docker containers
	docker-compose down

docker-logs: ## Show docker logs
	docker-compose logs -f

#install-tools: ## Install development tools
#	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
#	go install golang.org/x/tools/cmd/goimports@latest
#	go install github.com/golang/mock/mockgen@latest
#	go install github.com/golang/protobuf/protoc-gen-go@latest
#	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
#	go install github.com/cosmtrek/air@latest
#	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
#	go install gotest.tools/gotestsum@latest