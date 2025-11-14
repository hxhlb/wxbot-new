SHELL := /bin/bash

# 仅构建 32 位 Windows 可执行文件
OS := windows
ARCH := 386

APP := wxbot
MAIN := .

DIST := dist
OUT_EXE := $(DIST)/$(APP).exe

.PHONY: all build clean tidy run help example-basic example-auto-reply

all: build

help:
	@echo "可用命令:"
	@echo "  make build              - 构建 32 位 Windows 可执行文件"
	@echo "  make clean              - 清理 dist 目录"
	@echo "  make tidy               - 整理 Go 依赖"
	@echo "  make run                - 直接运行程序（需在 Windows 上）"
	@echo "  make example-basic      - 构建基本使用示例"
	@echo "  make example-auto-reply - 构建自动回复示例"

tidy:
	go mod tidy

build: tidy
	@mkdir -p $(DIST)
	GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(OUT_EXE) $(MAIN)
	@echo "[OK] 构建完成: $(OUT_EXE)"

build-debug: tidy
	@mkdir -p $(DIST)
	GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 go build -trimpath -o $(OUT_EXE) $(MAIN)
	@echo "[OK] 调试版本构建完成: $(OUT_EXE)"

example-basic: tidy
	@mkdir -p $(DIST)
	GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(DIST)/basic_usage.exe ./examples/basic_usage.go
	@echo "[OK] 基本示例构建完成: $(DIST)/basic_usage.exe"

example-auto-reply: tidy
	@mkdir -p $(DIST)
	GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(DIST)/auto_reply.exe ./examples/auto_reply.go
	@echo "[OK] 自动回复示例构建完成: $(DIST)/auto_reply.exe"

run:
	@if [ "$(shell go env GOOS)" != "windows" ]; then \
		echo "[错误] 此程序只能在 Windows 上运行"; \
		exit 1; \
	fi
	GOARCH=$(ARCH) go run $(MAIN)

clean:
	rm -rf $(DIST)
	@echo "[OK] 已清理 dist 目录"

# 格式化代码
fmt:
	go fmt ./...
	@echo "[OK] 代码格式化完成"

# 代码检查
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		go vet ./...; \
	fi
	@echo "[OK] 代码检查完成"

