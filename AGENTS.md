# Repository Guidelines

本指南面向贡献者，帮助你快速理解结构与开发流程，保持代码一致、稳定、可维护。

## 项目结构与模块组织
```
wxbot-new/
├─ main.go                 # 程序入口
├─ internal/
│  ├─ loader/              # DLL 加载与回调
│  ├─ memory/              # 共享内存管理
│  ├─ message/             # 消息类型与常量
│  └─ service/             # 业务编排与对外服务
├─ dist/                   # 构建产物（Windows .exe）
├─ Makefile                # 构建与工具命令
├─ go.mod / go.sum         # 依赖与版本
└─ pythondemo.py           # 对照/演示脚本（不参与构建）
```

## 构建、测试与本地开发命令
- `make tidy`：整理依赖（等价 `go mod tidy`）。
- `make build`：构建 32 位 Windows 可执行文件到 `dist/wxbot.exe`。
- `make build-debug`：调试构建（保留符号）。
- `make run`：在 Windows 上直接运行（需 32 位环境）。
- `make fmt`：格式化代码（`go fmt ./...`）。
- `make lint`：静态检查（优先 `golangci-lint`，否则 `go vet`）。
- `make clean`：清理 `dist/`。
- 示例：`make example-basic`、`make example-auto-reply` 构建示例程序。

必备文件：将 `NoveLoader.dll`、`NoveHelper.dll` 放在项目根目录。

## 代码风格与命名约定
- Go 官方风格：使用 `go fmt`；缩进用 tab；import 有序分组。
- 包名：全小写、简短、无下划线（如 `service`、`loader`）。
- 文件名：小写（可使用下划线），示例 `shared_memory.go`。
- 命名：导出标识符用 `PascalCase`，非导出用 `camelCase`。
- 错误处理：返回 `error` 并就地处理或向上返回；日志统一使用 `log`。

## 测试规范
- 框架：Go 原生 `testing`；文件放置于对应包的 `*_test.go`。
- 命名：`TestXxx(t *testing.T)`；表驱动测试优先。
- 运行：`go test ./... -v -cover`。
- 目标：新增功能配套单测，核心模块覆盖率 ≥ 60%。

## Commit 与 Pull Request
- 提交信息遵循 Conventional Commits：`feat(core): ...`、`fix(service): ...`、`docs: ...`、`refactor: ...`、`chore: ...`。
- PR 要求：
  - 说明动机、变更点与影响面；关联 Issue（如 `Closes #123`）。
  - 附本地验证结果（日志/截图）及风险说明。
  - 通过 `make fmt`、`make lint`、可构建；必要时附回滚方案。

## 安全与配置提示
- 仅支持 Windows 32 位：`GOOS=windows`、`GOARCH=386`。
- DLL 注入具侵入性，仅在授权与测试环境使用；勿提交 DLL 与敏感信息。
- 若被安全软件拦截，请在隔离/白名单环境中进行开发与验证。

