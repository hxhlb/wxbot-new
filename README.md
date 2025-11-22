# WxBot New

基于 DLL 注入技术的微信 (**版本 4.1.2.17**) 机器人服务，提供 HTTP API 接口进行微信自动化控制。  
接口文档（示例）：  
- Apifox: [https://s.apifox.cn/ea510d91-eb57-498a-924c-c35a2e9c1ea5](https://s.apifox.cn/ea510d91-eb57-498a-924c-c35a2e9c1ea5)  
- ShowDoc: [DLL调用接口文档 - ShowDoc](https://www.showdoc.com.cn/2447538212104511) （密码: `qqq222..`）

**请自行根据实际情况准备 `NoveLoader.dll`、`NoveHelper.dll` 注入文件**   
**本项目会有封号风险，自行承担后果**

## 项目特性

- ✅ **完整 HTTP API**：RESTful 设计，覆盖登录、消息、好友/群管理等能力
- ✅ **异步响应管理**：支持同步等待 DLL 回调结果，超时自动清理
- ✅ **自动重连**：心跳监控 + 自动重连机制，提升稳定性
- ✅ **配置管理**：支持动态更新配置、HTTP Basic 认证、日志/回调开关
- ✅ **消息回调推送**：可将指定消息类型转发至自定义 HTTP 回调地址
- ✅ **日志落盘**：按天滚动写入 `./logs/wxbot-YYYY-MM-DD.log`，同时输出控制台
- ✅ **安全混淆**：手动映射 DLL + 动态 Windows API 调用 + CGO 回调，降低静态检测风险

## 项目结构

```
wxbot-new/
├── main.go                     # 程序入口
├── config.example.json         # 配置示例
├── internal/
│   ├── api/                    # HTTP API 服务层
│   ├── config/                 # 配置管理
│   ├── loader/                 # DLL 加载 + 回调
│   ├── logging/                # 日志初始化（按天切割）
│   ├── memory/                 # 共享内存管理
│   ├── message/                # 消息类型与数据结构
│   ├── obfuscate/              # 混淆与动态 API 调用
│   ├── service/                # 微信业务服务层
│   └── utils/                  # 日志工具等
├── dist/                       # 构建产物（wxbot.exe）
├── Makefile                    # 构建脚本（推荐通过 make 构建）
├── pythondemo.py               # 旧 Python 示例（仅作对照，不参与构建）
├── NoveLoader.dll              # 加载器 DLL（需自行准备，运行时放在 exe 同目录）
└── NoveHelper.dll              # 助手 DLL（需自行准备，运行时放在 exe 同目录）
```

## 核心功能模块

### 1. HTTP API 服务层（`internal/api`）

提供完整的 RESTful API 接口：

#### 微信服务状态 / 登录

- `GET /api/wechat/status`          - 检查微信服务运行状态、当前 ClientID、连接数
- `GET /api/wechat/login-info`      - 获取当前登录账号信息（同步等待 DLL 回调）
- `POST /api/wechat/logout`         - 注销当前微信账号
- `GET /api/wechat/refresh-qrcode`  - 刷新登录二维码（同步）

#### 消息发送
> 半协议半Hook

- `POST /api/wechat/send-text`      - 发送普通文本消息（JSON）
- `POST /api/wechat/send-at-text`   - 发送群 @ 文本消息（支持 @ 全体）
- `POST /api/wechat/send-image`     - 发送图片消息（`multipart/form-data`，字段：`wxid` + `file`）
- `POST /api/wechat/send-file`      - 发送文件消息（`multipart/form-data`，字段：`wxid` + `file`）
- `POST /api/wechat/send-card`      - 发送名片消息

#### 好友 / 群管理

- `GET /api/wechat/friend-list`         - 获取好友列表
- `GET /api/wechat/friend-info`         - 获取指定好友信息（`?wxid=xxx`）
- `GET /api/wechat/group-list`          - 获取群列表（支持 `detail=0|1`，1 时包含成员列表）
- `GET /api/wechat/group-member-list`   - 获取群成员列表（`?room_wxid=xxx`）
- `POST /api/wechat/invite-group-member`- 邀请好友进群
- `POST /api/wechat/modify-group-name`  - 修改群名称

#### 小程序 / 语音

- `GET /api/wechat/mini-program-code`   - 获取小程序 `code` 及会话信息（`?appid=xxx`）
- `GET /api/wechat/voice-to-text`       - 语音转文本（`?msgid=xxx`）

#### 配置管理接口

- `GET /api/config`         - 查询完整配置
- `POST /api/config`        - 创建配置
- `PUT /api/config`         - 更新完整配置
- `DELETE /api/config`      - 删除配置（恢复默认）
- `GET /api/config/{key}`   - 查询单项配置
- `PUT /api/config/{key}`   - 更新单项配置

#### 健康检查

- `GET /health`             - 服务健康检查（不依赖微信 DLL）

#### 中间件

- **日志中间件**：记录所有 HTTP 请求（方法、路径、状态码、耗时）
- **CORS 中间件**：统一跨域配置
- **恢复中间件**：捕获 `panic` 并返回 500
- **内容类型中间件**：校验 `Content-Type`，仅允许 `application/json` / `multipart/form-data`
- **Basic 认证中间件**：当配置中存在 `auth` 时启用 HTTP Basic 认证
- **微信服务中间件**：对 `/api/wechat/*`（除 `/status`）自动检查微信服务是否已启动

### 2. 配置管理（`internal/config`）

- 支持 JSON 配置文件（默认 `config.json`，首次运行自动生成）
- 动态更新配置无需重启（通过 HTTP 接口修改后实时生效）
- 支持 HTTP Basic 认证用户列表
- 支持微信消息回调地址配置
- 线程安全的配置读写
- 默认监听 `0.0.0.0:5000`

配置字段说明：

- `host`：HTTP 服务监听地址，默认 `"0.0.0.0"`
- `port`：HTTP 服务端口，默认 `5000`
- `auth`：HTTP Basic 认证用户列表（为空则关闭认证）
- `log_recv_callback`：是否打印 DLL 回调消息日志（`0` 关闭，`1` 开启，默认 `1`）
- `callback_urls`：消息回调地址数组，当非空时会向这些地址推送部分消息

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
  ],
  "log_recv_callback": 1,
  "callback_urls": [
    "http://127.0.0.1:8080/wx/callback"
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

#### 消息回调推送

- 使用 `callback_urls` 配置一组 HTTP 回调地址
- 当收到特定消息类型（如聊天消息、登录/登出、好友/群事件等，范围大致为 `11046-11054` 及部分扩展码）时：
  - 以 `POST application/json` 向每个地址推送一条记录
  - 请求体格式：

```json
{
  "client_id": 12345,
  "msg_type": 11046,
  "data": { }
}
```

其中 `data` 字段结构可参考 `internal/message/types.go` 中相关数据结构定义。


### 4. DLL 加载层 (internal/loader)

#### DLL 函数调用（`loader.go` + `manual_map_windows.go`）

- 通过**硬编码偏移地址**调用 DLL 非导出函数
- 使用手动映射（Manual DLL Mapping）和动态 Windows API 调用加载 `NoveLoader.dll`
- 使用 `syscall.Syscall9` 进行底层调用
- 支持的操作：
  - 初始化微信 Socket
  - 注入微信进程（多种方式：普通、PID、多开）
  - 发送数据到微信
  - 销毁连接

#### 回调系统（`callback.go`）

- **连接回调**：客户端连接时触发
- **接收消息回调**：收到微信消息时触发
- **断开回调**：客户端断开时触发
- 使用 CGO 提供纯 C 回调函数，DLL 侧只看到标准 C 函数指针
- 自动解析 JSON 并统一转换为 `map[string]interface{}` 传给上层
- 线程安全的回调链管理（支持多回调同时注册）


## 快速开始

### 环境要求
- Go 1.18+
- 构建环境：macOS / Linux / Windows，需支持交叉编译
- 目标运行环境：32 位 Windows
- MinGW-w64 交叉编译器：`i686-w64-mingw32-gcc`（macOS/Linux）
- 运行时依赖：`NoveLoader.dll`、`NoveHelper.dll` 及对应 VC 运行库（`vcruntime140.dll`、`msvcp140.dll` 等，需与 exe 放在同一目录）

### 编译运行

```bash
# 使用 Makefile 编译（推荐）
make build

# 手动编译（32 位 Windows，启用 CGO）
GOOS=windows GOARCH=386 CGO_ENABLED=1 CC=i686-w64-mingw32-gcc go build -o dist/wxbot.exe .

# 运行（在 Windows 上）
cd dist
wxbot.exe
```

### 启动流程

1. 首次运行自动生成 `config.json` 配置文件（如不存在）
2. 创建共享内存 `windows_shell_global__`，写入固定密钥（供 DLL/微信进程探测）
3. 经过 2~6 秒随机延迟后开始执行注入逻辑（降低被检测概率）
4. 启动 HTTP API 服务，默认监听 `http://0.0.0.0:5000`
5. 初始化微信服务、加载 DLL、注册回调、启动心跳与自动重连协程

### HTTP API 使用示例

#### 检查服务状态
```bash
curl http://localhost:5000/api/wechat/status
```

响应：
```json
{
  "code": 0,
  "message": "状态检查完成",
  "data": {
    "running": true,
    "message": "微信服务正常运行",
    "client_id": 12345,
    "connected_count": 1
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

推荐通过配置 `callback_urls` 将消息推送给自建 HTTP 服务，在外部实现业务逻辑。

1. 在 `config.json` 中添加回调地址：

```json
{
  "callback_urls": [
    "http://127.0.0.1:8080/wx/callback"
  ]
}
```

2. 当收到聊天消息、登录/登出等事件时，WxBot 会向上述地址发送 `POST` 请求：

```json
{
  "client_id": 12345,
  "msg_type": 11046,
  "data": {
    "...": "具体字段见 internal/message/types.go"
  }
}
```

3. 在你的服务中根据 `msg_type` 和 `data` 进行路由与处理，例如收到聊天消息后自动回复、转发到 IM/队列等。

如需在进程内直接处理消息，可按需修改 `internal/service/service.go` 中的 `registerCallbacks()`，在现有回调逻辑中插入自定义处理代码。

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
    → 加载 DLl
    → 注册回调函数
    → 初始化 Socket
    → 注入 DLL
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

1. **仅支持 Windows 系统**：DLL 是 32 位的，必须编译为 32 位程序（`GOARCH=386`）
2. **DLL 文件**：需要自行准备 `vcruntime140.dll` `msvcp140.dll` (注意这是变种并非真实文件名)
3. **微信版本兼容性**：确保 DLL 与微信版本兼容，偏移地址硬编码
4. **安全性警告**：DLL 注入属于侵入性操作，请在授权环境下使用
5. **仅供学习**：本项目仅供学习交流，请勿用于非法用途
6. **更多 API 文档**：详见顶部 Apifox/ShowDoc 链接

## 常见问题

**Q: 编译后无法运行？**
A: 确保编译为 32 位（`GOARCH=386`），且 DLL 文件在程序同目录

**Q: 如何启用调试日志？**
A: 在代码中调用 `loader.SetDebugMode(true)` 可打印 DLL 回调的原始 JSON 数据（默认仅输出解析后的结构化日志）

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
