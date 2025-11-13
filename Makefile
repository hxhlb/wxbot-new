SHELL := /bin/sh

# 仅打包 32 位 Windows
OS := windows
ARCH := 386

APP := wxsvc
MAIN := ./cmd/server

DIST := dist
OUT_EXE := $(DIST)/$(APP).exe
STAGE := $(DIST)/$(APP)-$(OS)-$(ARCH)
ZIP := $(DIST)/$(APP)-$(OS)-$(ARCH).zip

.PHONY: all build package clean tidy

all: package

tidy:
	go mod tidy

build: tidy
	@mkdir -p $(DIST)
	GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 go build -trimpath -o $(OUT_EXE) $(MAIN)
	@echo "[OK] 构建完成: $(OUT_EXE)"

package: build
	@rm -rf $(STAGE)
	@mkdir -p $(STAGE)
	# 复制可执行文件
	cp -f $(OUT_EXE) $(STAGE)/
	# 可选复制 DLL（若存在）
	@if [ -f NoveLoader.dll ]; then cp -f NoveLoader.dll $(STAGE)/; fi
	@if [ -f NoveHelper.dll ]; then cp -f NoveHelper.dll $(STAGE)/; fi
	# 附带 README
	cp -f README.md $(STAGE)/ 2>/dev/null || true
	# 打包为 zip（需要 zip 工具）
	@cd $(DIST) && zip -r -q $(notdir $(ZIP)) $(notdir $(STAGE))
	@echo "[OK] 打包完成: $(ZIP)"

clean:
	rm -rf $(DIST)
	@echo "[OK] 已清理 dist 目录"

