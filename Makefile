# 分布式游戏排行榜系统 Makefile
# Author: HHaou
# Created: 2024-01-20

.PHONY: help build run test clean docker-build docker-run docker-stop deps lint fmt vet check

# 默认目标
help:
	@echo "分布式游戏排行榜系统 - 可用命令:"
	@echo "  build        - 构建应用程序"
	@echo "  run          - 运行应用程序"
	@echo "  test         - 运行测试"
	@echo "  clean        - 清理构建文件"
	@echo "  deps         - 下载依赖"
	@echo "  lint         - 运行代码检查"
	@echo "  fmt          - 格式化代码"
	@echo "  vet          - 运行go vet"
	@echo "  check        - 运行所有检查（fmt, vet, lint）"
	@echo "  docker-build - 构建Docker镜像"
	@echo "  docker-run   - 运行Docker容器"
	@echo "  docker-stop  - 停止Docker容器"
	@echo "  dev          - 启动开发环境"
	@echo "  dev-stop     - 停止开发环境"
	@echo "  logs         - 查看应用日志"

# 应用信息
APP_NAME := ranking-server
VERSION := 1.0.0
BUILD_TIME := $(shell date +%Y-%m-%dT%H:%M:%S)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建标志
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -w -s"
BUILD_FLAGS := -a -installsuffix cgo

# 构建应用程序
build:
	@echo "构建应用程序..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) $(BUILD_FLAGS) -o $(APP_NAME) ./cmd/server
	@echo "构建完成: $(APP_NAME)"

# 构建Windows版本
build-windows:
	@echo "构建Windows版本..."
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) $(BUILD_FLAGS) -o $(APP_NAME).exe ./cmd/server
	@echo "构建完成: $(APP_NAME).exe"

# 构建macOS版本
build-darwin:
	@echo "构建macOS版本..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) $(BUILD_FLAGS) -o $(APP_NAME)-darwin ./cmd/server
	@echo "构建完成: $(APP_NAME)-darwin"

# 构建所有平台版本
build-all: build build-windows build-darwin

# 运行应用程序
run:
	@echo "启动应用程序..."
	go run ./cmd/server --config configs/config.yaml

# 运行测试
test:
	@echo "运行测试..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "测试完成，覆盖率报告: coverage.html"

# 运行基准测试
bench:
	@echo "运行基准测试..."
	go test -bench=. -benchmem ./...

# 清理构建文件
clean:
	@echo "清理构建文件..."
	rm -f $(APP_NAME) $(APP_NAME).exe $(APP_NAME)-darwin
	rm -f coverage.out coverage.html
	@echo "清理完成"

# 下载依赖
deps:
	@echo "下载依赖..."
	go mod download
	go mod tidy
	@echo "依赖下载完成"

# 代码格式化
fmt:
	@echo "格式化代码..."
	go fmt ./...
	@echo "代码格式化完成"

# 运行go vet
vet:
	@echo "运行go vet..."
	go vet ./...
	@echo "go vet检查完成"

# 运行代码检查
lint:
	@echo "运行代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint未安装，跳过lint检查"; \
	fi

# 运行所有检查
check: fmt vet lint
	@echo "所有检查完成"

# 构建Docker镜像
docker-build:
	@echo "构建Docker镜像..."
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest
	@echo "Docker镜像构建完成"

# 运行Docker容器
docker-run:
	@echo "启动Docker容器..."
	docker-compose up -d
	@echo "Docker容器启动完成"

# 停止Docker容器
docker-stop:
	@echo "停止Docker容器..."
	docker-compose down
	@echo "Docker容器已停止"

# 启动开发环境
dev:
	@echo "启动开发环境..."
	docker-compose up -d mongodb redis
	@echo "等待数据库启动..."
	sleep 10
	@echo "开发环境启动完成"
	@echo "MongoDB: http://localhost:27017"
	@echo "Redis: localhost:6379"

# 停止开发环境
dev-stop:
	@echo "停止开发环境..."
	docker-compose down
	@echo "开发环境已停止"

# 启动完整环境（包含监控）
dev-full:
	@echo "启动完整开发环境..."
	docker-compose --profile tools --profile monitoring up -d
	@echo "完整开发环境启动完成"
	@echo "应用服务: http://localhost:8080"
	@echo "MongoDB管理: http://localhost:8081 (admin/admin)"
	@echo "Redis管理: http://localhost:8082 (admin/admin)"
	@echo "Prometheus: http://localhost:9091"
	@echo "Grafana: http://localhost:3000 (admin/admin)"

# 查看应用日志
logs:
	@echo "查看应用日志..."
	docker-compose logs -f ranking-service

# 查看所有服务日志
logs-all:
	@echo "查看所有服务日志..."
	docker-compose logs -f

# 重启应用服务
restart:
	@echo "重启应用服务..."
	docker-compose restart ranking-service
	@echo "应用服务已重启"

# 查看服务状态
status:
	@echo "查看服务状态..."
	docker-compose ps

# 进入应用容器
shell:
	@echo "进入应用容器..."
	docker-compose exec ranking-service sh

# 数据库备份
backup:
	@echo "备份MongoDB数据..."
	mkdir -p backups
	docker-compose exec mongodb mongodump --db ranking --out /tmp/backup
	docker cp $$(docker-compose ps -q mongodb):/tmp/backup ./backups/mongodb-$$(date +%Y%m%d-%H%M%S)
	@echo "数据备份完成"

# 数据库恢复
restore:
	@echo "恢复MongoDB数据..."
	@echo "请指定备份目录: make restore BACKUP_DIR=backups/mongodb-20240120-120000"
	@if [ -z "$(BACKUP_DIR)" ]; then \
		echo "错误: 请指定BACKUP_DIR参数"; \
		exit 1; \
	fi
	docker cp $(BACKUP_DIR) $$(docker-compose ps -q mongodb):/tmp/restore
	docker-compose exec mongodb mongorestore --db ranking --drop /tmp/restore/ranking
	@echo "数据恢复完成"

# 性能测试
load-test:
	@echo "运行性能测试..."
	@if command -v hey >/dev/null 2>&1; then \
		echo "测试健康检查接口..."; \
		hey -n 1000 -c 10 http://localhost:8080/health; \
		echo "测试排行榜查询接口..."; \
		hey -n 1000 -c 10 http://localhost:8080/api/v1/leaderboard/global_score_demo; \
	else \
		echo "hey工具未安装，请先安装: go install github.com/rakyll/hey@latest"; \
	fi

# 安装开发工具
install-tools:
	@echo "安装开发工具..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/rakyll/hey@latest
	@echo "开发工具安装完成"

# 生成API文档
docs:
	@echo "生成API文档..."
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/server/main.go -o docs; \
	else \
		echo "swag工具未安装，请先安装: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

# 版本信息
version:
	@echo "应用版本: $(VERSION)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "Git提交: $(GIT_COMMIT)"