.PHONY: help build run test clean docker-up docker-down docker-logs docker-ps install lint fmt vet tidy bench

# 变量定义
APP_NAME := go-next-erp
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GO_VERSION := $(shell go version | awk '{print $$3}')

# 构建标志
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Docker 相关
DOCKER_COMPOSE := docker-compose
DOCKER_COMPOSE_FILE := docker-compose.yml

# 颜色输出
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

## help: 显示帮助信息
help:
	@echo ''
	@echo '使用方法:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo '目标:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "  ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

## 开发相关:

## install: 安装依赖
install:
	@echo "${GREEN}Installing dependencies...${RESET}"
	go mod download
	go mod verify

## tidy: 整理依赖
tidy:
	@echo "${GREEN}Tidying dependencies...${RESET}"
	go mod tidy

## fmt: 格式化代码
fmt:
	@echo "${GREEN}Formatting code...${RESET}"
	go fmt ./...

## vet: 代码静态检查
vet:
	@echo "${GREEN}Running go vet...${RESET}"
	go vet ./...

## lint: 代码规范检查（需要 golangci-lint）
lint:
	@echo "${GREEN}Running golangci-lint...${RESET}"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "${YELLOW}golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest${RESET}"; \
	fi

## 构建相关:

## build: 编译项目
build:
	@echo "${GREEN}Building $(APP_NAME) $(VERSION)...${RESET}"
	@echo "  Go Version: $(GO_VERSION)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  Build Time: $(BUILD_TIME)"
	go build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/server

## build-linux: 编译 Linux 版本
build-linux:
	@echo "${GREEN}Building $(APP_NAME) for Linux...${RESET}"
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-linux-amd64 ./cmd/server

## build-darwin: 编译 macOS 版本
build-darwin:
	@echo "${GREEN}Building $(APP_NAME) for macOS...${RESET}"
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-darwin-amd64 ./cmd/server
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(APP_NAME)-darwin-arm64 ./cmd/server

## build-windows: 编译 Windows 版本
build-windows:
	@echo "${GREEN}Building $(APP_NAME) for Windows...${RESET}"
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-windows-amd64.exe ./cmd/server

## build-all: 编译所有平台版本
build-all: build-linux build-darwin build-windows
	@echo "${GREEN}All builds completed!${RESET}"

## 运行相关:

## run: 运行应用
run: build
	@echo "${GREEN}Running $(APP_NAME)...${RESET}"
	./bin/$(APP_NAME)

## dev: 开发模式运行（热重载需要 air）
dev:
	@echo "${GREEN}Running in development mode...${RESET}"
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "${YELLOW}air not installed. Installing...${RESET}"; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

## 测试相关:

## test: 运行所有测试
test:
	@echo "${GREEN}Running tests...${RESET}"
	go test -v -race -coverprofile=coverage.out ./...

## test-short: 运行短测试（跳过集成测试）
test-short:
	@echo "${GREEN}Running short tests...${RESET}"
	go test -v -short ./...

## test-coverage: 运行测试并生成覆盖率报告
test-coverage: test
	@echo "${GREEN}Generating coverage report...${RESET}"
	go tool cover -html=coverage.out -o coverage.html
	@echo "${CYAN}Coverage report: coverage.html${RESET}"

## test-unit: 运行单元测试
test-unit:
	@echo "${GREEN}Running unit tests...${RESET}"
	go test -v -race ./pkg/...

## bench: 运行基准测试
bench:
	@echo "${GREEN}Running benchmarks...${RESET}"
	go test -bench=. -benchmem -run=^$$ ./...

## bench-logger: 运行日志基准测试
bench-logger:
	@echo "${GREEN}Running logger benchmarks...${RESET}"
	go test -bench=. -benchmem -run=^$$ ./pkg/logger

## Docker 相关:

## docker-build: 构建 Docker 镜像
docker-build:
	@echo "${GREEN}Building Docker image...${RESET}"
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest

## docker-up: 启动 Docker Compose 服务
docker-up:
	@echo "${GREEN}Starting Docker Compose services...${RESET}"
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) up -d
	@echo "${CYAN}Services started. Use 'make docker-ps' to check status.${RESET}"

## docker-down: 停止 Docker Compose 服务
docker-down:
	@echo "${GREEN}Stopping Docker Compose services...${RESET}"
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) down

## docker-restart: 重启 Docker Compose 服务
docker-restart: docker-down docker-up

## docker-ps: 查看 Docker 容器状态
docker-ps:
	@echo "${GREEN}Docker containers status:${RESET}"
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) ps

## docker-logs: 查看所有容器日志
docker-logs:
	@echo "${GREEN}Showing all container logs...${RESET}"
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) logs -f

## docker-logs-postgres: 查看 PostgreSQL 容器日志
docker-logs-postgres:
	@echo "${GREEN}Showing erp-postgres container logs...${RESET}"
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) logs -f erp-postgres

## docker-logs-redis: 查看 Redis 容器日志
docker-logs-redis:
	@echo "${GREEN}Showing erp-redis container logs...${RESET}"
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) logs -f erp-redis

## docker-logs-minio: 查看 MinIO 容器日志
docker-logs-minio:
	@echo "${GREEN}Showing erp-minio container logs...${RESET}"
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) logs -f erp-minio

## docker-logs-milvus: 查看 Milvus 容器日志
docker-logs-milvus:
	@echo "${GREEN}Showing erp-milvus container logs...${RESET}"
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) logs -f erp-milvus

## docker-logs-milvus-etcd: 查看 Milvus etcd 容器日志
docker-logs-milvus-etcd:
	@echo "${GREEN}Showing erp-milvus-etcd container logs...${RESET}"
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) logs -f erp-milvus-etcd

## docker-exec-postgres: 进入 PostgreSQL 容器
docker-exec-postgres:
	@echo "${GREEN}Entering erp-postgres container...${RESET}"
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) exec erp-postgres psql -U postgres -d erp

## docker-exec-redis: 进入 Redis 容器
docker-exec-redis:
	@echo "${GREEN}Entering erp-redis container...${RESET}"
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) exec erp-redis redis-cli -a redis123

## docker-exec-minio: 进入 MinIO 容器
docker-exec-minio:
	@echo "${GREEN}Entering erp-minio container...${RESET}"
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) exec erp-minio sh

## 清理相关:

## clean: 清理构建文件
clean:
	@echo "${GREEN}Cleaning build files...${RESET}"
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache -testcache

## clean-docker: 清理 Docker 资源
clean-docker:
	@echo "${GREEN}Cleaning Docker resources...${RESET}"
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) down -v --remove-orphans
	docker system prune -f

## clean-all: 清理所有文件（包括依赖）
clean-all: clean clean-docker
	@echo "${GREEN}Cleaning all files...${RESET}"
	rm -rf vendor/
	go clean -modcache

## Protobuf 相关:

## proto-install: 安装 Protobuf 工具
proto-install:
	@echo "${GREEN}Installing Protobuf tools...${RESET}"
	@if ! command -v buf >/dev/null 2>&1; then \
		echo "${YELLOW}Installing buf...${RESET}"; \
		go install github.com/bufbuild/buf/cmd/buf@latest; \
	else \
		echo "${CYAN}buf already installed${RESET}"; \
	fi
	@if ! command -v protoc-gen-go >/dev/null 2>&1; then \
		echo "${YELLOW}Installing protoc-gen-go...${RESET}"; \
		go install google.golang.org/protobuf/cmd/protoc-gen-go@latest; \
	else \
		echo "${CYAN}protoc-gen-go already installed${RESET}"; \
	fi
	@if ! command -v protoc-gen-go-grpc >/dev/null 2>&1; then \
		echo "${YELLOW}Installing protoc-gen-go-grpc...${RESET}"; \
		go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest; \
	else \
		echo "${CYAN}protoc-gen-go-grpc already installed${RESET}"; \
	fi
	@if ! command -v wire >/dev/null 2>&1; then \
		echo "${YELLOW}Installing wire...${RESET}"; \
		go install github.com/google/wire/cmd/wire@latest; \
	else \
		echo "${CYAN}wire already installed${RESET}"; \
	fi

## proto-gen: 生成 Protobuf 代码
proto-gen:
	@echo "${GREEN}Generating Protobuf code...${RESET}"
	@if command -v buf >/dev/null 2>&1; then \
		buf generate; \
	else \
		echo "${YELLOW}buf not installed. Run: make proto-install${RESET}"; \
	fi

## proto-lint: 检查 Protobuf 代码规范
proto-lint:
	@echo "${GREEN}Linting Protobuf files...${RESET}"
	@if command -v buf >/dev/null 2>&1; then \
		buf lint; \
	else \
		echo "${YELLOW}buf not installed. Run: make proto-install${RESET}"; \
	fi

## proto-breaking: 检查 Protobuf 破坏性变更
proto-breaking:
	@echo "${GREEN}Checking for breaking changes in Protobuf...${RESET}"
	@if command -v buf >/dev/null 2>&1; then \
		buf breaking --against '.git#branch=main'; \
	else \
		echo "${YELLOW}buf not installed. Run: make proto-install${RESET}"; \
	fi

## wire-gen: 生成 Wire 依赖注入代码
wire-gen:
	@echo "${GREEN}Generating Wire code...${RESET}"
	@if command -v wire >/dev/null 2>&1; then \
		cd cmd/server && wire; \
	else \
		echo "${YELLOW}wire not installed. Run: make proto-install${RESET}"; \
	fi

## 工具相关:

## tools-install: 安装开发工具
tools-install:
	@echo "${GREEN}Installing development tools...${RESET}"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/cosmtrek/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest

## swag-init: 生成 Swagger 文档
swag-init:
	@echo "${GREEN}Generating Swagger documentation...${RESET}"
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/server/main.go; \
	else \
		echo "${YELLOW}swag not installed. Run: make tools-install${RESET}"; \
	fi

## 数据库相关:

## db-migrate-up: 运行数据库迁移（需要 migrate 工具）
db-migrate-up:
	@echo "${GREEN}Running database migrations...${RESET}"
	@if command -v migrate >/dev/null 2>&1; then \
		migrate -path ./migrations -database "postgresql://postgres:password@localhost:5432/erp?sslmode=disable" up; \
	else \
		echo "${YELLOW}migrate not installed. Visit: https://github.com/golang-migrate/migrate${RESET}"; \
	fi

## db-migrate-down: 回滚数据库迁移
db-migrate-down:
	@echo "${GREEN}Rolling back database migrations...${RESET}"
	@if command -v migrate >/dev/null 2>&1; then \
		migrate -path ./migrations -database "postgresql://postgres:password@localhost:5432/erp?sslmode=disable" down; \
	else \
		echo "${YELLOW}migrate not installed. Visit: https://github.com/golang-migrate/migrate${RESET}"; \
	fi

## 信息相关:

## version: 显示版本信息
version:
	@echo "${CYAN}$(APP_NAME) Version Information:${RESET}"
	@echo "  Version:    $(VERSION)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Go Version: $(GO_VERSION)"

## info: 显示项目信息
info:
	@echo "${CYAN}Project Information:${RESET}"
	@echo "  App Name:   $(APP_NAME)"
	@echo "  Version:    $(VERSION)"
	@echo "  Go Version: $(GO_VERSION)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "${CYAN}Dependencies:${RESET}"
	@go list -m all | head -10

## 快捷命令:

## all: 格式化、检查、测试、构建
all: fmt vet test build
	@echo "${GREEN}All tasks completed!${RESET}"

## ci: CI/CD 流程（格式检查、测试、构建）
ci: fmt vet lint test-coverage build
	@echo "${GREEN}CI tasks completed!${RESET}"
