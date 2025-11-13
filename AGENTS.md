# Repository Guidelines

本仓库为基于 Python `pythondemo.py` 的 Go 语言重构版本，聚焦跨平台（Windows/非 Windows）运行与易维护架构。

## 项目结构与模块
- `cmd/server/`：程序入口（启动服务与 HTTP API）。
- `internal/wxsvc/`：核心服务层（Loader 接口、Windows/非 Windows 实现、回调分发、业务逻辑）。
- `internal/httpapi/`：HTTP 路由（`/send` 等）。
- `internal/logger/`：日志初始化（文件 + 控制台）。
- `pythondemo.py`：原始 Python 脚本（对照参考）。
- 根目录：`README.md`、`Makefile`、`go.mod`。

## 构建、运行与开发
- 本机运行（非 Windows 为 stub 调试）：
  - `go run ./cmd/server`
- 仅构建（32 位 Windows 可执行）：
  - `make build`（产物：`dist/wxsvc.exe`，等价 `GOOS=windows GOARCH=386 go build -o dist/wxsvc.exe ./cmd/server`）
- 常用环境变量：`API_HOST`、`API_PORT`、`LOADER_PATH`、`DLL_PATH`。

## 代码风格与命名
- Go 标准格式化（gofmt），缩进使用 tab；提交前执行 `go mod tidy`。
- 包名小写、无下划线；导出标识使用驼峰（如 `NewService`）。
- 目录按职责分层（cmd/internal）；避免循环依赖，公共类型放入 `internal/wxsvc`。
- 日志统一使用 `internal/logger`；错误优先返回 `error`，不吞异常。

## 测试规范
- 目前暂无测试；新增测试使用 Go 原生测试框架：
  - 文件命名：`xxx_test.go`，与被测代码同目录。
  - 运行：`go test ./...`，建议增加关键路径覆盖。

## 提交与 PR 规范
- Commit 信息建议：`type(scope): short description`（如 `feat(wxsvc): add reconnect logic`）。
- PR 描述包含：变更目的、核心改动、验证方式（命令/日志片段）、相关 issue。
- 避免混入无关重构；保持小步提交，便于代码审查。

## 安全与配置提示
- Windows 环境建议 32 位（`GOARCH=386`）；需要 `NoveLoader.dll`、`NoveHelper.dll` 与偏移匹配。
- 非 Windows 使用 stub，不做真实注入，仅用于接口联调。
- 不要提交密钥/敏感 DLL；路径与偏移更新请在 `internal/wxsvc/loader_windows.go` 内集中维护。
