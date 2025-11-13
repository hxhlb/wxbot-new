# wxbot-new (Go 版本)

基于 `pythondemo.py` 的 Go 重构，提供跨平台可编译与运行的服务结构：

- 非 Windows 平台：使用 stub 实现，服务可运行并提供 HTTP API，消息发送会记录日志。
- Windows 平台：已实现基于 DLL 基址+偏移的调用（与 Python 版一致，默认要求 32 位 DLL）。

## 目录结构

```
cmd/server/main.go          # 入口，启动服务与 HTTP API
internal/logger/logger.go   # 日志初始化（文件+控制台）
internal/httpapi/httpapi.go # HTTP /send 接口
internal/wxsvc/types.go     # 常量定义
internal/wxsvc/loader.go    # 底层 Loader 接口
internal/wxsvc/loader_stub.go     # 非 Windows 的 stub 实现
internal/wxsvc/loader_windows.go  # Windows 实现（偏移调用）
internal/wxsvc/service.go   # 服务管理、重连、发送逻辑
internal/wxsvc/handler.go   # 回调接口（OnConnect/OnReceive/OnClose）
internal/wxsvc/handler_default.go # 默认回调处理器（登录时发送启动消息等）
```

## 运行

```
go run ./cmd/server
```

环境变量：

- `API_HOST`（默认 `0.0.0.0`）
- `API_PORT`（默认 `5000`）
- `LOADER_PATH`（默认 `./NoveLoader.dll`）
- `DLL_PATH`（默认 `./NoveHelper.dll`）

启动后访问：`POST /send`

示例 1：自定义 payload

```
curl -X POST http://localhost:5000/send \
  -H 'Content-Type: application/json' \
  -d '{"type":11036, "data":{"room_wxid":"47945916190@chatroom","content":"hello"}}'
```

示例 2：文本直发

```
curl -X POST http://localhost:5000/send \
  -H 'Content-Type: application/json' \
  -d '{"text":"hello","room_wxid":"47945916190@chatroom"}'
```

## 备注
 
 - Windows 与 Python 版一致，推荐 32 位环境（GOARCH=386）。若使用 64 位 Go，请确保 DLL 与偏移匹配，否则将无法注入。
 - Windows 构建（32 位）：
   - `GOOS=windows GOARCH=386 go build -o wxsvc.exe ./cmd/server`
 - 非 Windows 平台使用 stub，不做实际注入，仅用于开发与 API 联调。

## 注册自定义回调

实现 `internal/wxsvc/handler.go` 的 `Handler` 接口，并在启动时注册：

```
type MyHandler struct{}

func (MyHandler) OnConnect(clientID uint32) {}
func (MyHandler) OnReceive(clientID uint32, typ int, data map[string]any) {}
func (MyHandler) OnClose(clientID uint32) {}

// 在 cmd/server/main.go 中：
svc := wxsvc.NewService(loaderPath, dllPath)
svc.AddHandler(MyHandler{})
```

默认已内置一个回调处理器，会在用户登录后发送一次启动消息，并记录连接数与日志。

## Loader 接口（与 Python 版能力对齐）

- InitWeChatSocket(connect, recv, close) bool
- InjectWeChat(dllPath) (clientID, error)
- SendWeChatData(clientID, message) (bool, error)
- DestroyWeChat() error
- UseUtf8() bool
- InjectWeChat2(dllPath, exePath) (clientID, error)
- InjectWeChatPid(pid, dllPath) (clientID, error)
- InjectWeChatMultiOpen(dllPath, exePath) (clientID, error)
- GetUserWeChatVersion() (string, error)
- GetInstallWeixinVersion() (string, error)  // 若偏移未知返回空串
