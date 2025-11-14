# API 架构设计文档

## 架构概览

项目采用分层架构,清晰分离职责,便于扩展和维护。

```
internal/api/
├── server.go           # HTTP 服务器管理
├── router.go           # 路由注册
├── middleware.go       # 中间件集合
├── response.go         # 统一响应工具
├── config_handler.go   # 配置接口处理器
└── [future]_handler.go # 未来的其他处理器
```

## 模块职责

### 1. server.go - 服务器管理层
**职责**: HTTP 服务器生命周期管理

```go
// 创建服务器
server := api.NewServer(host, port, configManager)

// 启动服务
server.Start()

// 停止服务
server.Stop()
```

**特性**:
- 优雅关闭(5秒超时)
- 请求超时配置(读15s/写15s)
- 自动集成路由和中间件

---

### 2. router.go - 路由注册层
**职责**: 集中管理所有路由规则

```go
type Router struct {
    mux           *http.ServeMux
    configHandler *ConfigHandler
    // 未来扩展: userHandler, messageHandler 等
}

func (r *Router) RegisterRoutes() http.Handler {
    // 注册路由
    r.mux.HandleFunc("/api/config", r.handleConfig)

    // 应用中间件链
    return Chain(
        RecoveryMiddleware(),
        LoggingMiddleware(),
        CORSMiddleware(),
        ContentTypeMiddleware(),
    )(r.mux)
}
```

**设计优势**:
- 路由集中管理,一目了然
- 通过方法路由(handleConfig)统一处理 CRUD
- 易于添加新的 Handler

**扩展示例**:
```go
// 未来添加用户管理接口
type Router struct {
    configHandler *ConfigHandler
    userHandler   *UserHandler    // 新增
}

func (r *Router) RegisterRoutes() http.Handler {
    // 配置管理
    r.mux.HandleFunc("/api/config", r.handleConfig)

    // 用户管理(新增)
    r.mux.HandleFunc("/api/users", r.handleUsers)
    r.mux.HandleFunc("/api/users/", r.handleUserItem)

    return Chain(...)(r.mux)
}
```

---

### 3. middleware.go - 中间件层
**职责**: 提供可复用的请求处理逻辑

**已实现中间件**:

#### LoggingMiddleware - 日志记录
```go
// 功能: 记录每个请求的方法、路径、状态码、耗时
[HTTP] GET /api/config - Status: 200 - 耗时: 2.5ms
```

#### CORSMiddleware - 跨域支持
```go
// 允许所有源跨域访问
// 支持方法: GET, POST, PUT, DELETE, OPTIONS
// 允许头: Content-Type, Authorization
```

#### RecoveryMiddleware - panic 恢复
```go
// 捕获 panic,防止服务崩溃
// 自动返回 500 错误
```

#### ContentTypeMiddleware - 内容类型检查
```go
// POST/PUT 请求必须是 application/json
// 否则返回 415 Unsupported Media Type
```

**中间件链设计**:
```go
// 从外到内执行顺序:
Recovery -> Logging -> CORS -> ContentType -> Handler
```

**扩展示例**:
```go
// 添加认证中间件
func AuthMiddleware(requiredRole string) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := r.Header.Get("Authorization")
            if !validateToken(token, requiredRole) {
                ErrorResponse(w, http.StatusUnauthorized, "未授权")
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// 使用
handler := Chain(
    RecoveryMiddleware(),
    LoggingMiddleware(),
    AuthMiddleware("admin"),  // 新增
    CORSMiddleware(),
)(r.mux)
```

---

### 4. response.go - 响应工具层
**职责**: 统一 HTTP 响应格式

**统一响应结构**:
```go
type Response struct {
    Code    int         `json:"code"`    // 0=成功, 其他=失败
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

**工具函数**:
```go
// 成功响应
SuccessResponse(w, "操作成功", data)
// => {"code":0, "message":"操作成功", "data":{...}}

// 错误响应
ErrorResponse(w, 400, "参数错误")
// => {"code":400, "message":"参数错误"}

// 自定义响应
JSONResponse(w, 200, Response{...})
```

---

### 5. config_handler.go - 配置处理器
**职责**: 处理配置相关的业务逻辑

**接口列表**:
| 方法 | 路径 | Handler 方法 |
|------|------|--------------|
| GET | `/api/config` | GetConfig |
| POST | `/api/config` | CreateConfig |
| PUT | `/api/config` | UpdateConfig |
| DELETE | `/api/config` | DeleteConfig |
| GET | `/api/config/{key}` | GetConfigItem |
| PUT | `/api/config/{key}` | UpdateConfigItem |

**Handler 结构**:
```go
type ConfigHandler struct {
    configManager *config.Manager
}

// 每个方法负责单一职责
func (h *ConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request)
func (h *ConfigHandler) UpdateConfig(w http.ResponseWriter, r *http.Request)
// ...
```

**扩展示例**:
```go
// 添加微信消息处理器
type MessageHandler struct {
    messageService *service.MessageService
}

func NewMessageHandler(svc *service.MessageService) *MessageHandler {
    return &MessageHandler{messageService: svc}
}

func (h *MessageHandler) SendText(w http.ResponseWriter, r *http.Request) {
    // 解析请求
    var req SendTextRequest
    json.NewDecoder(r.Body).Decode(&req)

    // 调用服务
    if err := h.messageService.SendText(req.ClientID, req.Wxid, req.Text); err != nil {
        ErrorResponse(w, 500, err.Error())
        return
    }

    SuccessResponse(w, "发送成功", nil)
}
```

---

## 请求处理流程

```
客户端请求
    ↓
HTTP Server (server.go)
    ↓
中间件链 (middleware.go)
    ├─ RecoveryMiddleware (捕获 panic)
    ├─ LoggingMiddleware (记录日志)
    ├─ CORSMiddleware (处理跨域)
    └─ ContentTypeMiddleware (检查内容类型)
    ↓
路由匹配 (router.go)
    ↓
Handler 处理 (config_handler.go)
    ├─ 参数验证
    ├─ 调用业务逻辑 (config.Manager)
    └─ 响应返回 (response.go)
    ↓
返回客户端
```

---

## 扩展指南

### 1. 添加新的接口模块

**步骤1**: 创建 Handler 文件
```go
// internal/api/user_handler.go
package api

type UserHandler struct {
    userService *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
    return &UserHandler{userService: svc}
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
    users := h.userService.GetAll()
    SuccessResponse(w, "查询成功", users)
}
```

**步骤2**: 在 router.go 中注册
```go
type Router struct {
    configHandler *ConfigHandler
    userHandler   *UserHandler  // 新增
}

func NewRouter(cm *config.Manager, us *service.UserService) *Router {
    return &Router{
        configHandler: NewConfigHandler(cm),
        userHandler:   NewUserHandler(us),  // 新增
    }
}

func (r *Router) RegisterRoutes() http.Handler {
    // 原有路由
    r.mux.HandleFunc("/api/config", r.handleConfig)

    // 新增路由
    r.mux.HandleFunc("/api/users", r.handleUsers)

    return Chain(...)(r.mux)
}

func (r *Router) handleUsers(w http.ResponseWriter, req *http.Request) {
    switch req.Method {
    case http.MethodGet:
        r.userHandler.ListUsers(w, req)
    case http.MethodPost:
        r.userHandler.CreateUser(w, req)
    default:
        ErrorResponse(w, 405, "不支持的方法")
    }
}
```

**步骤3**: 更新 server.go (如需注入新依赖)
```go
func NewServer(host string, port int, cm *config.Manager, us *service.UserService) *Server {
    return &Server{
        host:   host,
        port:   port,
        router: NewRouter(cm, us),
    }
}
```

### 2. 添加新的中间件

```go
// internal/api/middleware.go

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(rps int) Middleware {
    limiter := rate.NewLimiter(rate.Limit(rps), rps)

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                ErrorResponse(w, 429, "请求过于频繁")
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// 在 router.go 中使用
handler := Chain(
    RecoveryMiddleware(),
    LoggingMiddleware(),
    RateLimitMiddleware(100),  // 限制 100 req/s
    CORSMiddleware(),
)(r.mux)
```

### 3. 自定义响应格式

如需支持分页、元数据等:
```go
// response.go
type PaginatedResponse struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
    Meta    struct {
        Page       int `json:"page"`
        PageSize   int `json:"page_size"`
        TotalCount int `json:"total_count"`
    } `json:"meta"`
}

func PaginatedSuccessResponse(w http.ResponseWriter, data interface{}, page, size, total int) {
    resp := PaginatedResponse{
        Code:    0,
        Message: "查询成功",
        Data:    data,
    }
    resp.Meta.Page = page
    resp.Meta.PageSize = size
    resp.Meta.TotalCount = total

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}
```

---

## 设计原则

1. **单一职责**: 每个文件/模块只负责一件事
2. **开闭原则**: 对扩展开放,对修改关闭
3. **依赖注入**: Handler 通过构造函数接收依赖
4. **分层清晰**: Server → Router → Middleware → Handler
5. **统一规范**: 响应格式、错误处理、日志格式统一

---

## 优势总结

✅ **易扩展**: 添加新接口只需创建 Handler,在 Router 注册
✅ **易维护**: 每个模块职责清晰,修改影响范围小
✅ **易测试**: Handler 可独立测试,中间件可单独测试
✅ **代码复用**: 中间件、响应工具全局复用
✅ **团队协作**: 不同开发者可并行开发不同 Handler

---

## 常见场景示例

### 场景1: 添加微信消息发送接口

```go
// 1. 创建 message_handler.go
type MessageHandler struct {
    wechatService *service.WeChatService
}

func (h *MessageHandler) SendText(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Wxid string `json:"wxid"`
        Text string `json:"text"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    if err := h.wechatService.SendTextMessage(req.Wxid, req.Text); err != nil {
        ErrorResponse(w, 500, err.Error())
        return
    }

    SuccessResponse(w, "发送成功", nil)
}

// 2. 在 router.go 注册
r.mux.HandleFunc("/api/message/send", r.messageHandler.SendText)
```

### 场景2: 添加认证鉴权

```go
// 1. 在 middleware.go 添加
func JWTMiddleware() Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := r.Header.Get("Authorization")
            if !validateJWT(token) {
                ErrorResponse(w, 401, "未授权")
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// 2. 在 router.go 应用
handler := Chain(
    RecoveryMiddleware(),
    LoggingMiddleware(),
    JWTMiddleware(),  // 添加到链中
    CORSMiddleware(),
)(r.mux)
```

---

## 性能优化建议

1. **连接池**: 使用 http.Server 的 IdleTimeout
2. **超时控制**: 设置合理的读写超时
3. **限流**: 使用 RateLimitMiddleware
4. **缓存**: Handler 中可集成缓存层
5. **异步处理**: 耗时操作使用 goroutine

```go
func (h *MessageHandler) SendBulkMessages(w http.ResponseWriter, r *http.Request) {
    var req BulkMessageRequest
    json.NewDecoder(r.Body).Decode(&req)

    // 异步处理
    go func() {
        for _, msg := range req.Messages {
            h.wechatService.SendTextMessage(msg.Wxid, msg.Text)
        }
    }()

    SuccessResponse(w, "批量发送任务已提交", nil)
}
```
