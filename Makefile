SHELL := /bin/sh

# 仅构建 32 位 Windows 可执行文件
OS := windows
ARCH := 386

APP := wxsvc
MAIN := ./cmd/server

DIST := dist
OUT_EXE := $(DIST)/$(APP).exe

.PHONY: all build clean tidy

all: build

tidy:
	go mod tidy

build: tidy
	@mkdir -p $(DIST)
	GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 go build -trimpath -o $(OUT_EXE) $(MAIN)
	@echo "[OK] 构建完成: $(OUT_EXE)"

clean:
	rm -rf $(DIST)
	@echo "[OK] 已清理 dist 目录"

