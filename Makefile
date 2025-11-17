SHELL := /bin/bash

# 仅构建 32 位 Windows 可执行文件
OS := windows
ARCH := 386

APP := wxbot
MAIN := .

DIST := dist
OUT_EXE := $(DIST)/$(APP).exe

# DLL 源目录和嵌入目录
LIBS_DIR := libs
EMBEDDED_DIR := internal/loader/embedded

.PHONY: all build clean tidy run help example-basic example-auto-reply prepare-embed

all: build

help:
	@echo "可用命令:"
	@echo "  make build              - 构建 32 位 Windows 可执行文件 (DLL已嵌入)"
	@echo "  make build-classic      - 构建不嵌入DLL的版本 (需要外部DLL文件)"
	@echo "  make clean              - 清理 dist 目录"
	@echo "  make tidy               - 整理 Go 依赖"
	@echo "  make run                - 直接运行程序（需在 Windows 上）"
	@echo "  make example-basic      - 构建基本使用示例"
	@echo "  make example-auto-reply - 构建自动回复示例"

tidy:
	go mod tidy

# 准备嵌入资源
prepare-embed:
	@echo "[INFO] 准备嵌入DLL资源..."
	@mkdir -p $(EMBEDDED_DIR)
	@if [ -f "$(LIBS_DIR)/NoveLoader.dll" ]; then \
		cp $(LIBS_DIR)/NoveLoader.dll $(EMBEDDED_DIR)/; \
		echo "[OK] 已复制 NoveLoader.dll"; \
	else \
		echo "[WARN] 未找到 NoveLoader.dll"; \
	fi
	@if [ -f "$(LIBS_DIR)/NoveHelper.dll" ]; then \
		cp $(LIBS_DIR)/NoveHelper.dll $(EMBEDDED_DIR)/; \
		echo "[OK] 已复制 NoveHelper.dll"; \
	else \
		echo "[WARN] 未找到 NoveHelper.dll"; \
	fi

# 构建 (嵌入DLL)
build: tidy prepare-embed
	@mkdir -p $(DIST)
	GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(OUT_EXE) $(MAIN)
	@echo "[OK] 构建完成 (DLL已嵌入): $(OUT_EXE)"
	@ls -lh $(OUT_EXE)

# 构建经典版本 (不嵌入DLL,需要外部文件)
build-classic: tidy
	@mkdir -p $(DIST)
	GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -tags=noembed -o $(OUT_EXE) $(MAIN)
	@echo "[OK] 构建完成 (需要外部DLL): $(OUT_EXE)"
	@echo "[INFO] 请确保运行时目录包含 NoveLoader.dll 和 NoveHelper.dll"

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

