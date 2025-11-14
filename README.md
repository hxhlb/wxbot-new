# WxBot New

基于 DLL 注入技术的微信(**版本 4.1.2.17**)机器人服务，提供 HTTP API 接口进行微信自动化操作。   
接口文档: [https://s.apifox.cn/ea510d91-eb57-498a-924c-c35a2e9c1ea5](https://s.apifox.cn/ea510d91-eb57-498a-924c-c35a2e9c1ea5)

## 项目特性

- ✅ **完整的 HTTP API**：RESTful API 设计，易于集成
- ✅ **异步响应管理**：支持同步等待响应的消息操作
- ✅ **自动重连**：心跳监控 + 自动重连机制
- ✅ **配置管理**：支持动态配置和 HTTP Basic 认证
- ✅ **高性能**：Go 实现，内存占用低，并发能力强
- ✅ **生产就绪**：包含中间件、日志、优雅关闭等企业级特性

## 项目结构

```
wxbot-new/
├── main.go                           # 程序入口
├── config.json                       # 配置文件（自动生成）
├── internal/
│   ├── api/                          # HTTP API 服务层
│   │   ├── server.go                 # HTTP 服务器管理
│   │   ├── router.go                 # 路由注册
│   │   ├── wechat_handler.go         # 微信接口处理
│   │   ├── config_handler.go         # 配置接口处理
│   │   ├── middleware.go             # 中间件（日志、CORS、认证等）
│   │   └── response.go               # 统一响应格式
│   ├── config/                       # 配置管理
│   │   └── config.go                 # 配置加载、更新、保存
│   ├── service/                      # 微信业务服务层
│   │   ├── service.go                # 服务生命周期管理
│   │   ├── response_manager.go       # 异步响应管理器
│   │   ├── message.go                # 消息发送方法
│   │   ├── user.go                   # 用户信息获取
│   │   └── contact.go                # 联系人管理
│   ├── loader/                       # DLL 加载层
│   │   ├── loader.go                 # DLL 加载和函数调用
│   │   └── callback.go               # Go ↔ C 回调转换
│   ├── message/                      # 消息类型定义
│   │   └── types.go                  # 消息常量和数据结构
│   └── memory/                       # 共享内存管理
│       └── shared_memory.go          # Windows 共享内存操作
├── NoveLoader.dll                    # 加载器DLL（需自行准备）
└── NoveHelper.dll                    # 助手DLL（需自行准备）
```

## 核心功能模块

### 1. HTTP API 服务层 (internal/api)

提供完整的 RESTful API 接口：

#### 微信操作接口
- `GET /api/wechat/status` - 检查服务状态
- `GET /api/wechat/login-info` - 获取登录信息（同步）
- `GET /api/user-info` - 获取用户信息（别名）
- `GET /api/wechat/refresh-qrcode` - 刷新二维码（同步）

#### 配置管理接口
- `GET /api/config` - 查询完整配置
- `POST /api/config` - 创建配置
- `PUT /api/config` - 更新完整配置
- `DELETE /api/config` - 删除配置
- `GET /api/config/{key}` - 查询单项配置
- `PUT /api/config/{key}` - 更新单项配置

#### 健康检查
- `GET /health` - 服务健康检查

#### 中间件
- **日志中间件**：记录所有 HTTP 请求
- **CORS 中间件**：支持跨域请求
- **恢复中间件**：捕获 panic 并返回 500
- **内容类型中间件**：验证 Content-Type
- **Basic 认证中间件**：可选的 HTTP Basic 认证

### 2. 配置管理 (internal/config)

- 支持 JSON 配置文件（默认 `config.json`）
- 动态更新配置无需重启
- HTTP Basic 认证用户管理
- 线程安全的配置读写
- 默认监听 `0.0.0.0:5000`

配置示例：
```json
{
  "host": "0.0.0.0",
  "port": 5000,
  "auth": [
    {
      "username": "admin",
      "password": "password123"
    }
  ]
}
```

### 3. 微信业务服务层 (internal/service)

#### 服务管理 (service.go)
- **生命周期管理**：Initialize → Start → Stop
- **心跳监控**：每 60 秒更新心跳时间戳
- **自动重连**：120 秒无心跳自动重连（最多 5 次，间隔 10 秒）
- **连接状态管理**：追踪已连接的客户端

#### 异步响应管理 (response_manager.go)
- 支持同步等待响应的消息操作
- 自动超时处理（默认 10 秒）
- 定期清理过期请求
- 基于消息类型 + ClientID 匹配响应


### 4. DLL 加载层 (internal/loader)

#### DLL 函数调用 (loader.go)
- 通过**硬编码偏移地址**调用 DLL 非导出函数
- 使用 `syscall.Syscall9` 进行底层调用
- 支持的操作：
  - 初始化微信 Socket
  - 注入微信进程
  - 发送数据到微信
  - 销毁连接
  - 多种注入方式（PID 注入、多开等）

#### 回调系统 (callback.go)
- **连接回调**：客户端连接时触发
- **接收消息回调**：收到微信消息时触发
- **断开回调**：客户端断开时触发
- 使用 `windows.NewCallback` 实现 Go ↔ C 回调转换
- 自动解析 JSON 数据
- 线程安全的回调链管理


## 快速开始

### 环境要求
- Go 1.18+
- Windows 系统（32 位）
- NoveLoader.dll 和 NoveHelper.dll

### 编译运行

```bash
# 使用 Makefile 编译（推荐）
make build

# 手动编译（32 位 Windows）
GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -o dist/wxbot.exe main.go

# 运行
cd dist
wxbot.exe
```

### 启动流程

1. **首次运行**会自动生成 `config.json` 配置文件
2. **延迟 3 秒**后创建共享内存（给微信进程启动时间）
3. **HTTP API 服务**启动在 `http://0.0.0.0:5000`
4. **微信服务**自动初始化并注入 DLL

### HTTP API 使用示例

#### 检查服务状态
```bash
curl http://localhost:5000/api/wechat/status
```

响应：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "client_id": 12345,
    "connected_clients": 1,
    "is_running": true
  }
}
```

#### 获取登录信息
```bash
curl http://localhost:5000/api/wechat/login-info
```

响应：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "account": "your_account",
    "avatar": "https://...",
    "nickname": "Your Name",
    "wxid": "wxid_xxxxx"
  }
}
```

#### 刷新二维码
```bash
curl http://localhost:5000/api/wechat/refresh-qrcode
```

响应：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "file": "C:\\path\\to\\qrcode.png",
    "qrcode": "base64_encoded_qrcode_data",
    "pid": 12345
  }
}
```

#### 配置管理

查询配置：
```bash
curl http://localhost:5000/api/config
```

更新配置：
```bash
curl -X PUT http://localhost:5000/api/config \
  -H "Content-Type: application/json" \
  -d '{"host":"0.0.0.0","port":5000,"auth":[{"username":"admin","password":"123456"}]}'
```

更新单项配置：
```bash
curl -X PUT http://localhost:5000/api/config/port \
  -H "Content-Type: application/json" \
  -d '{"value":8080}'
```

### 自定义消息处理

修改 `internal/service/service.go` 中的 `registerCallbacks()` 方法：

```go
// 接收消息回调
s.loader.GetCallbackManager().AddRecvCallback(func(clientID uintptr, msgType int, data map[string]interface{}) {
    switch message.MessageType(msgType) {
    case message.MTChatMessage:
        // 处理聊天消息
        log.Printf("收到聊天消息: %v", data)

        // 自动回复示例
        content, _ := data["content"].(string)
        fromWxid, _ := data["fromWxid"].(string)
        if content == "ping" {
            s.HelperSendText(fromWxid, "pong")
        }

    case message.MTUserLogin:
        // 处理用户登录事件
        log.Printf("用户登录: %v", data)

    case message.MTUserLogout:
        // 处理用户登出事件
        log.Printf("用户登出: %v", data)
    }
})
```

## 技术架构

### 分层架构

```
┌─────────────────────────────────────┐
│         HTTP API 服务层              │ (api/)
│   - Server, Router, Handler          │
│   - Middleware (日志/CORS/认证)       │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│        微信业务服务层                 │ (service/)
│   - WeChatService                    │
│   - ResponseManager (异步响应管理)    │
│   - HelperSend*/HelperGet* 方法      │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│      DLL加载和回调转换层              │ (loader/)
│   - NoveLoader (DLL 函数调用)        │
│   - CallbackManager (Go ↔ C 回调)   │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│     基础设施层 + Windows API          │ (memory/, message/)
│   - SharedMemoryManager              │
│   - MessageType 定义                 │
│   - kernel32.dll, ntdll.dll API      │
└─────────────────────────────────────┘
```

### 数据流

#### 启动流程
```
main()
  → 加载配置 (config.json)
  → 创建共享内存 (延迟3秒)
  → 启动 HTTP API 服务 (0.0.0.0:5000)
  → 启动微信服务
    → 加载 NoveLoader.dll
    → 注册回调函数
    → 初始化 Socket
    → 注入 NoveHelper.dll
    → 启动心跳监控协程
    → 启动响应清理协程
```

#### 消息发送流程 (同步方法)
```
HTTP Request → WeChatHandler
  → WeChatService.HelperGetCurrentLoginInfo()
    → ResponseManager.RegisterRequest() (注册请求，返回 Channel)
    → SendMessage(JSON) → NoveLoader.SendWeChatData()
      → syscall.Syscall9(offsetSendWeChatData, ...)
    → ResponseManager.WaitForResponse() (阻塞等待)
      ← DLL 回调 onRecv()
      ← CallbackManager 触发 RecvCallback
      ← WeChatService 处理响应
      ← ResponseManager.HandleResponse() (写入 Channel)
    → 解析响应并返回
  ← HTTP Response (JSON)
```

#### 心跳和重连流程
```
runService() 主循环 (每1秒检查)
  → 检查 time.Since(lastHeartbeat) > 120秒?
    YES → reconnect()
      → DestroyWeChat()
      → sleep(10秒)
      → InjectWeChat() (重新注入)
      → 更新 lastHeartbeat

startHeartbeat() 协程 (每60秒)
  → 更新 lastHeartbeat = time.Now()
```

## 注意事项

1. **仅支持 Windows 32 位**：DLL 是 32 位的，必须编译为 32 位程序（`GOARCH=386`）
2. **DLL 文件**：需要自行准备 `NoveLoader.dll` 和 `NoveHelper.dll`
3. **微信版本兼容性**：确保 DLL 与微信版本兼容，偏移地址硬编码
4. **安全性警告**：DLL 注入属于侵入性操作，请在授权环境下使用
5. **仅供学习**：本项目仅供学习交流，请勿用于非法用途
6. **更多 API 文档**：https://www.showdoc.com.cn/2447538212104511 (密码: qqq222..)

## 常见问题

**Q: 编译后无法运行？**
A: 确保编译为 32 位（`GOARCH=386`），且 DLL 文件在程序同目录

**Q: 如何启用调试日志？**
A: 在代码中调用 `loader.SetDebugMode(true)`

**Q: HTTP API 如何认证？**
A: 在 `config.json` 中配置 `auth` 字段，启用 HTTP Basic 认证

**Q: 消息发送失败？**
A: 检查 ClientID 是否有效、微信是否已登录、网络是否正常

**Q: 心跳超时频繁？**
A: 可能是微信进程不稳定，检查微信版本兼容性

**Q: 如何扩展新接口？**
A: 在 `internal/api/router.go` 中注册新路由，在对应 Handler 中实现逻辑

## 声明

本项目仅供学习交流使用，请勿用于非法用途。使用本项目产生的任何法律责任由使用者自行承担。
