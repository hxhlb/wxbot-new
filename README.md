# wxbot-new (Go 版本)

基于 DLL 注入技术的微信机器人服务，Go 语言实现版本。

## 项目结构

```
wxbot-new/
├── main.go                    # 程序入口
├── internal/
│   ├── memory/               # 共享内存管理
│   │   └── shared_memory.go
│   ├── loader/               # DLL加载器
│   │   ├── loader.go         # DLL加载和函数调用
│   │   └── callback.go       # 回调系统
│   ├── message/              # 消息类型定义
│   │   └── types.go
│   └── service/              # 微信服务
│       ├── service.go        # 服务管理器
│       └── helper.go         # 消息发送助手
├── NoveLoader.dll            # 加载器DLL（需自行准备）
└── NoveHelper.dll            # 助手DLL（需自行准备）
```

## 核心功能模块

### 1. 共享内存管理 (internal/memory)
- 创建并写入 33 字节共享内存
- 共享内存名称: `windows_shell_global__`
- 固定密钥: `3101b223dca7715b0154924f0eeeee20`

### 2. DLL加载器 (internal/loader)
- 动态加载 NoveLoader.dll
- 通过偏移地址调用非导出函数
- 支持的操作：
  - 初始化微信Socket
  - 注入微信进程
  - 发送数据到微信
  - 销毁连接
  - 多种注入方式（PID注入、多开等）

### 3. 回调系统 (internal/loader)
- 连接回调：客户端连接时触发
- 接收消息回调：收到微信消息时触发
- 断开回调：客户端断开时触发

### 4. 消息类型 (internal/message)
支持的消息类型：
- `11024`: 调试日志
- `11025`: 用户登录
- `11026`: 用户登出
- `11030`: 好友列表
- `11036`: 发送文本消息
- `11037`: 发送@消息
- `11038`: 发送卡片
- `11039`: 发送链接
- `11040`: 发送图片
- `11041`: 发送文件
- `11042`: 发送视频
- `11043`: 发送GIF
- `11046`: 聊天消息

### 5. 服务管理器 (internal/service)
- 服务启动/停止/重连
- 心跳监控（60秒间隔）
- 连接超时检测（120秒）
- 自动重连机制（最多5次）
- 连接状态管理

### 6. 消息发送助手 (internal/service)
提供便捷的消息发送方法：
- `HelperGetFriendList()`: 获取好友列表
- `HelperSendText(toWxid, content)`: 发送文本
- `HelperSendAtText(toWxid, content, atList)`: 发送@消息
- `HelperSendCard(toWxid, cardWxid)`: 发送卡片
- `HelperSendURL(toWxid, title, desc, url, imageURL)`: 发送链接
- `HelperSendImage(toWxid, filePath)`: 发送图片
- `HelperSendFile(toWxid, filePath)`: 发送文件
- `HelperSendVideo(toWxid, filePath)`: 发送视频
- `HelperSendGif(toWxid, filePath)`: 发送GIF

## 使用方法

### 环境要求
- Go 1.21+
- Windows 系统（32位）
- NoveLoader.dll 和 NoveHelper.dll

### 编译运行

```bash
# 编译（32位）
set GOARCH=386
go build -o wxbot.exe main.go

# 运行
wxbot.exe
```

### 代码示例

```go
package main

import (
    "wxbot-new/internal/service"
    "wxbot-new/internal/memory"
)

func main() {
    // 1. 初始化共享内存
    memManager := memory.NewSharedMemoryManager()
    memManager.CreateAndWriteSharedMemory()
    defer memManager.Close()

    // 2. 创建微信服务
    wechatService := service.NewWeChatService(
        "./NoveLoader.dll",
        "./NoveHelper.dll",
    )

    // 3. 启动服务
    wechatService.Start()

    // 4. 发送消息示例
    // wechatService.HelperSendText("filehelper", "Hello from Go!")
}
```

### 自定义消息处理

修改 `internal/service/service.go` 中的 `registerCallbacks()` 方法：

```go
// 接收消息回调
s.loader.AddRecvCallback(func(clientID uintptr, msgType int, data map[string]interface{}) {
    switch message.MessageType(msgType) {
    case message.MTChatMessage:
        // 处理聊天消息
        log.Printf("收到聊天消息: %v", data)

        // 自动回复示例
        // s.HelperSendText("filehelper", "收到消息")
    }
})
```

## 与 Python 版本的对比

| 特性 | Python 版本 | Go 版本 |
|------|------------|---------|
| 依赖库 | ctypes, threading | syscall, unsafe |
| 性能 | 中等 | 更高 |
| 内存占用 | 较高 | 较低 |
| 并发模型 | GIL限制 | goroutine |
| 部署 | 需要Python环境 | 单一可执行文件 |
| 类型安全 | 动态类型 | 静态类型 |

## 注意事项

1. **仅支持 Windows 32位**：DLL 是 32 位的，需要编译为 32 位程序
2. **DLL 文件**：需要自行准备 `NoveLoader.dll` 和 `NoveHelper.dll`
3. **微信版本**：确保 DLL 与微信版本兼容
4. **安全性**：DLL 注入属于侵入性操作，请在授权环境下使用
5. **API文档**：更多 API 请参考 https://www.showdoc.com.cn/2447538212104511 (密码: qqq222..)

## 许可证

本项目仅供学习交流使用，请勿用于非法用途。

## 技术架构说明

### 从 Python 到 Go 的迁移要点

1. **共享内存操作**
   - Python: `ctypes.WinDLL + CreateFileMappingA`
   - Go: `syscall.LazyDLL + NewProc`

2. **回调函数**
   - Python: `@WINFUNCTYPE` 装饰器
   - Go: `syscall.NewCallback`

3. **DLL 函数调用**
   - Python: 通过偏移地址计算函数指针
   - Go: 通过 `syscall.Syscall9` 直接调用

4. **并发模型**
   - Python: `threading.Thread`
   - Go: `goroutine + channel`

5. **JSON 序列化**
   - Python: `json.dumps/loads`
   - Go: `encoding/json`

