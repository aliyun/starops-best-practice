# STAROps SDK Samples - Makefile
# STAROps SDK 示例项目 - 构建脚本
#
# This Makefile provides common commands for building, testing, and running
# the STAROps SDK sample programs.
# 本 Makefile 提供构建、测试和运行 STAROps SDK 示例程序的常用命令。
#
# Usage / 使用方法:
#   make help     - Show available commands / 显示可用命令
#   make build    - Build all sample programs / 构建所有示例程序
#   make test     - Run all tests / 运行所有测试
#   make run      - Run the default chat example / 运行默认对话示例
#   make lint     - Run linting checks / 运行代码检查
#   make clean    - Clean build artifacts / 清理构建产物

# Go parameters / Go 参数
GO := go
GOFLAGS := -v
GOFMT := gofmt
GOVET := $(GO) vet

# Project directories / 项目目录
CMD_DIR := cmd
BIN_DIR := bin

# Sample programs / 示例程序
# List of all sample programs to build
# 需要构建的所有示例程序列表
SAMPLES := \
	chat \
	chat-simple \
	chat-timeout \
	chat-interactive \
	chat-multi-turn \
	thread-manager \
	chat-from-file \
	chat-with-time

# Binary targets / 二进制目标
BINARIES := $(addprefix $(BIN_DIR)/,$(SAMPLES))

# Default target / 默认目标
.DEFAULT_GOAL := help

# ============================================================================
# Build targets / 构建目标
# ============================================================================

.PHONY: build
## Build all sample programs / 构建所有示例程序
build: $(BINARIES)
	@echo "✅ All sample programs built successfully / 所有示例程序构建成功"

# Build individual binaries / 构建单个二进制文件
$(BIN_DIR)/%: $(CMD_DIR)/%/main.go
	@echo "🔨 Building $*..."
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $@ ./$(<D)

.PHONY: build-chat
## Build chat example / 构建 chat 示例
build-chat: $(BIN_DIR)/chat

.PHONY: build-chat-simple
## Build chat-simple example / 构建 chat-simple 示例
build-chat-simple: $(BIN_DIR)/chat-simple

.PHONY: build-chat-timeout
## Build chat-timeout example / 构建 chat-timeout 示例
build-chat-timeout: $(BIN_DIR)/chat-timeout

.PHONY: build-chat-interactive
## Build chat-interactive example / 构建 chat-interactive 示例
build-chat-interactive: $(BIN_DIR)/chat-interactive

.PHONY: build-chat-multi-turn
## Build chat-multi-turn example / 构建 chat-multi-turn 示例
build-chat-multi-turn: $(BIN_DIR)/chat-multi-turn

.PHONY: build-thread-manager
## Build thread-manager example / 构建 thread-manager 示例
build-thread-manager: $(BIN_DIR)/thread-manager

.PHONY: build-chat-from-file
## Build chat-from-file example / 构建 chat-from-file 示例
build-chat-from-file: $(BIN_DIR)/chat-from-file

.PHONY: build-chat-with-time
## Build chat-with-time example / 构建 chat-with-time 示例
build-chat-with-time: $(BIN_DIR)/chat-with-time

# ============================================================================
# Test targets / 测试目标
# ============================================================================

.PHONY: test
## Run all tests / 运行所有测试
test:
	@echo "🧪 Running tests / 运行测试..."
	$(GO) test $(GOFLAGS) ./...
	@echo "✅ All tests passed / 所有测试通过"

.PHONY: test-coverage
## Run tests with coverage report / 运行测试并生成覆盖率报告
test-coverage:
	@echo "🧪 Running tests with coverage / 运行测试并生成覆盖率..."
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html / 覆盖率报告已生成: coverage.html"

.PHONY: test-verbose
## Run tests with verbose output / 运行测试（详细输出）
test-verbose:
	@echo "🧪 Running tests (verbose) / 运行测试（详细模式）..."
	$(GO) test -v ./...

# ============================================================================
# Run targets / 运行目标
# ============================================================================

.PHONY: run
## Run the default chat example / 运行默认对话示例
run: build-chat
	@echo "🚀 Running chat example / 运行 chat 示例..."
	@echo "💡 Make sure to set environment variables first / 请确保已设置环境变量"
	@echo "   See .env.example for required variables / 参见 .env.example 了解所需变量"
	./$(BIN_DIR)/chat

.PHONY: run-simple
## Run chat-simple example / 运行简洁模式示例
run-simple: build-chat-simple
	@echo "🚀 Running chat-simple example / 运行 chat-simple 示例..."
	./$(BIN_DIR)/chat-simple

.PHONY: run-timeout
## Run chat-timeout example / 运行超时控制示例
run-timeout: build-chat-timeout
	@echo "🚀 Running chat-timeout example / 运行 chat-timeout 示例..."
	./$(BIN_DIR)/chat-timeout

.PHONY: run-interactive
## Run chat-interactive example / 运行交互事件示例
run-interactive: build-chat-interactive
	@echo "🚀 Running chat-interactive example / 运行 chat-interactive 示例..."
	./$(BIN_DIR)/chat-interactive

.PHONY: run-multi-turn
## Run chat-multi-turn example / 运行多轮对话示例
run-multi-turn: build-chat-multi-turn
	@echo "🚀 Running chat-multi-turn example / 运行 chat-multi-turn 示例..."
	./$(BIN_DIR)/chat-multi-turn

.PHONY: run-thread-manager
## Run thread-manager example / 运行会话管理示例
run-thread-manager: build-thread-manager
	@echo "🚀 Running thread-manager example / 运行 thread-manager 示例..."
	./$(BIN_DIR)/thread-manager

.PHONY: run-from-file
## Run chat-from-file example / 运行文件加载示例
run-from-file: build-chat-from-file
	@echo "🚀 Running chat-from-file example / 运行 chat-from-file 示例..."
	./$(BIN_DIR)/chat-from-file

.PHONY: run-with-time
## Run chat-with-time example / 运行时间参数示例
run-with-time: build-chat-with-time
	@echo "🚀 Running chat-with-time example / 运行 chat-with-time 示例..."
	./$(BIN_DIR)/chat-with-time

# ============================================================================
# Lint targets / 代码检查目标
# ============================================================================

.PHONY: lint
## Run all linting checks / 运行所有代码检查
lint: vet fmt-check
	@echo "✅ All lint checks passed / 所有代码检查通过"

.PHONY: vet
## Run go vet / 运行 go vet 检查
vet:
	@echo "🔍 Running go vet / 运行 go vet..."
	$(GOVET) ./...

.PHONY: fmt-check
## Check code formatting / 检查代码格式
fmt-check:
	@echo "🔍 Checking code formatting / 检查代码格式..."
	@if [ -n "$$($(GOFMT) -l .)" ]; then \
		echo "❌ The following files need formatting / 以下文件需要格式化:"; \
		$(GOFMT) -l .; \
		echo "Run 'make fmt' to fix / 运行 'make fmt' 修复"; \
		exit 1; \
	fi
	@echo "✅ Code formatting is correct / 代码格式正确"

.PHONY: fmt
## Format code / 格式化代码
fmt:
	@echo "🔧 Formatting code / 格式化代码..."
	$(GOFMT) -w .
	@echo "✅ Code formatted / 代码已格式化"

# ============================================================================
# Clean targets / 清理目标
# ============================================================================

.PHONY: clean
## Clean build artifacts / 清理构建产物
clean:
	@echo "🧹 Cleaning build artifacts / 清理构建产物..."
	rm -rf $(BIN_DIR)
	rm -f coverage.out coverage.html
	@echo "✅ Clean complete / 清理完成"

# ============================================================================
# Dependency targets / 依赖管理目标
# ============================================================================

.PHONY: deps
## Download dependencies / 下载依赖
deps:
	@echo "📦 Downloading dependencies / 下载依赖..."
	$(GO) mod download
	@echo "✅ Dependencies downloaded / 依赖下载完成"

.PHONY: deps-tidy
## Tidy dependencies / 整理依赖
deps-tidy:
	@echo "📦 Tidying dependencies / 整理依赖..."
	$(GO) mod tidy
	@echo "✅ Dependencies tidied / 依赖整理完成"

.PHONY: deps-verify
## Verify dependencies / 验证依赖
deps-verify:
	@echo "📦 Verifying dependencies / 验证依赖..."
	$(GO) mod verify
	@echo "✅ Dependencies verified / 依赖验证完成"

# ============================================================================
# Development targets / 开发目标
# ============================================================================

.PHONY: dev
## Run development checks (lint + test) / 运行开发检查（代码检查 + 测试）
dev: lint test
	@echo "✅ Development checks passed / 开发检查通过"

.PHONY: all
## Build and test all / 构建并测试所有
all: deps lint build test
	@echo "✅ All tasks completed / 所有任务完成"

# ============================================================================
# Help target / 帮助目标
# ============================================================================

.PHONY: help
## Show this help message / 显示帮助信息
help:
	@echo ""
	@echo "STAROps SDK Samples - Makefile Commands"
	@echo "STAROps SDK 示例项目 - Makefile 命令"
	@echo ""
	@echo "Usage / 使用方法: make [target]"
	@echo ""
	@echo "Available targets / 可用目标:"
	@echo ""
	@echo "  Build / 构建:"
	@echo "    build              Build all sample programs / 构建所有示例程序"
	@echo "    build-chat         Build chat example / 构建 chat 示例"
	@echo "    build-chat-simple  Build chat-simple example / 构建 chat-simple 示例"
	@echo "    build-chat-timeout Build chat-timeout example / 构建 chat-timeout 示例"
	@echo "    build-chat-interactive Build chat-interactive example / 构建 chat-interactive 示例"
	@echo "    build-chat-multi-turn Build chat-multi-turn example / 构建 chat-multi-turn 示例"
	@echo "    build-thread-manager Build thread-manager example / 构建 thread-manager 示例"
	@echo "    build-chat-from-file Build chat-from-file example / 构建 chat-from-file 示例"
	@echo "    build-chat-with-time Build chat-with-time example / 构建 chat-with-time 示例"
	@echo ""
	@echo "  Test / 测试:"
	@echo "    test               Run all tests / 运行所有测试"
	@echo "    test-coverage      Run tests with coverage / 运行测试并生成覆盖率"
	@echo "    test-verbose       Run tests with verbose output / 详细模式运行测试"
	@echo ""
	@echo "  Run / 运行:"
	@echo "    run                Run default chat example / 运行默认 chat 示例"
	@echo "    run-simple         Run chat-simple example / 运行 chat-simple 示例"
	@echo "    run-timeout        Run chat-timeout example / 运行 chat-timeout 示例"
	@echo "    run-interactive    Run chat-interactive example / 运行 chat-interactive 示例"
	@echo "    run-multi-turn     Run chat-multi-turn example / 运行 chat-multi-turn 示例"
	@echo "    run-thread-manager Run thread-manager example / 运行 thread-manager 示例"
	@echo "    run-from-file      Run chat-from-file example / 运行 chat-from-file 示例"
	@echo "    run-with-time      Run chat-with-time example / 运行 chat-with-time 示例"
	@echo ""
	@echo "  Lint / 代码检查:"
	@echo "    lint               Run all lint checks / 运行所有代码检查"
	@echo "    vet                Run go vet / 运行 go vet"
	@echo "    fmt-check          Check code formatting / 检查代码格式"
	@echo "    fmt                Format code / 格式化代码"
	@echo ""
	@echo "  Clean / 清理:"
	@echo "    clean              Clean build artifacts / 清理构建产物"
	@echo ""
	@echo "  Dependencies / 依赖:"
	@echo "    deps               Download dependencies / 下载依赖"
	@echo "    deps-tidy          Tidy dependencies / 整理依赖"
	@echo "    deps-verify        Verify dependencies / 验证依赖"
	@echo ""
	@echo "  Development / 开发:"
	@echo "    dev                Run lint + test / 运行代码检查和测试"
	@echo "    all                Build and test all / 构建并测试所有"
	@echo "    help               Show this help / 显示帮助"
	@echo ""
	@echo "Examples / 示例:"
	@echo "  make build           # Build all programs / 构建所有程序"
	@echo "  make test            # Run all tests / 运行所有测试"
	@echo "  make run             # Run chat example / 运行 chat 示例"
	@echo "  make lint            # Check code quality / 检查代码质量"
	@echo "  make clean           # Clean up / 清理"
	@echo ""
