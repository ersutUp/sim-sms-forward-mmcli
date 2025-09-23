# Makefile for sim-sms-forward

# 项目信息
PROJECT_NAME := sim-sms-forward
VERSION := v1.0.0
BINARY_NAME := sim-sms-forward
OUTPUT_DIR := dist

# 构建信息
BUILD_TIME := $(shell date '+%Y-%m-%d %H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 编译标志
LDFLAGS := -s -w
LDFLAGS += -X 'main.Version=$(VERSION)'
LDFLAGS += -X 'main.BuildTime=$(BUILD_TIME)'
LDFLAGS += -X 'main.GitCommit=$(GIT_COMMIT)'

# 默认目标
.PHONY: all
all: build

# 本地构建
.PHONY: build
build:
	@echo "构建本地版本..."
	go build -ldflags="$(LDFLAGS)" -o $(BINARY_NAME) main.go

# 清理
.PHONY: clean
clean:
	@echo "清理构建文件..."
	rm -rf $(OUTPUT_DIR)
	rm -f $(BINARY_NAME)

# 跨平台构建
.PHONY: build-all
build-all: clean
	@echo "开始跨平台构建..."
	@./build.sh

# 快速构建主要平台
.PHONY: build-main
build-main: clean
	@echo "构建主要平台..."
	@mkdir -p $(OUTPUT_DIR)
	# Linux amd64
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64 main.go
	# Linux arm64
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-arm64 main.go
	# Windows amd64
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)-windows-amd64.exe main.go
	# macOS amd64
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-amd64 main.go
	# macOS arm64
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-arm64 main.go
	@echo "主要平台构建完成！"

# 代码格式化
.PHONY: fmt
fmt:
	@echo "格式化代码..."
	go fmt ./...

# 代码检查
.PHONY: vet
vet:
	@echo "代码检查..."
	go vet ./...

# 依赖整理
.PHONY: tidy
tidy:
	@echo "整理依赖..."
	go mod tidy

# 测试
.PHONY: test
test:
	@echo "运行测试..."
	go test ./...

# 安装依赖
.PHONY: deps
deps:
	@echo "安装依赖..."
	go mod download

# 开发环境准备
.PHONY: dev-setup
dev-setup: deps fmt vet tidy
	@echo "开发环境准备完成！"

# 发布准备
.PHONY: release
release: dev-setup test build-all
	@echo "发布包准备完成！"

# 运行 (需要配置文件)
.PHONY: run
run: build
	@if [ -f config.json ]; then \
		./$(BINARY_NAME) config.json; \
	else \
		echo "请先创建 config.json 配置文件"; \
		echo "可以复制 config.example.json 并修改"; \
	fi

# 帮助信息
.PHONY: help
help:
	@echo "可用的构建命令:"
	@echo "  build        - 构建本地版本"
	@echo "  build-all    - 跨平台构建所有版本"
	@echo "  build-main   - 构建主要平台版本"
	@echo "  clean        - 清理构建文件"
	@echo "  fmt          - 代码格式化"
	@echo "  vet          - 代码检查"
	@echo "  tidy         - 整理依赖"
	@echo "  test         - 运行测试"
	@echo "  dev-setup    - 开发环境准备"
	@echo "  release      - 发布准备"
	@echo "  run          - 运行程序"
	@echo "  help         - 显示此帮助信息"