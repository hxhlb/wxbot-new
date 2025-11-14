# HTTP API 接口文档

## 概述

微信机器人提供 RESTful API 用于管理配置文件,支持配置的查询、创建、修改、删除操作。

**基础信息**
- 默认地址: `http://0.0.0.0:5000`
- 配置文件: `config.json`
- 响应格式: JSON
- 认证方式: HTTP Basic Authentication (可选)

## 响应格式

所有接口返回统一格式:

```json
{
  "code": 0,              // 0=成功, 其他=失败
  "message": "操作描述",
  "data": {}              // 响应数据(可选)
}
```

## 认证说明

### HTTP Basic Authentication

如果配置文件中设置了 `auth` 数组,所有 API 请求都需要提供 HTTP Basic 认证。

**配置示例**
```json
{
  "host": "0.0.0.0",
  "port": 5000,
  "auth": [
    {
      "username": "admin",
      "password": "secret123"
    }
  ]
}
```

**使用认证**
```bash
# 方式1: 使用 -u 参数
curl -u admin:secret123 http://localhost:5000/api/config

# 方式2: 手动设置 Authorization 头
curl -H "Authorization: Basic YWRtaW46c2VjcmV0MTIz" http://localhost:5000/api/config
```

**注意事项**
- 如果 `auth` 数组为空或不存在,则不启用认证
- 认证失败返回 401 状态码
- 支持多个用户账号

---

## 接口列表

### 1. 查询完整配置

**请求**
```http
GET /api/config
```

**响应示例**
```json
{
  "code": 0,
  "message": "查询成功",
  "data": {
    "host": "0.0.0.0",
    "port": 5000,
    "auth": [
      {
        "username": "admin",
        "password": "secret123"
      }
    ]
  }
}
```

**cURL 示例**
```bash
# 无认证
curl http://localhost:5000/api/config

# 带认证
curl -u admin:secret123 http://localhost:5000/api/config
```

---

### 2. 创建/更新完整配置

**请求**
```http
POST /api/config/create
Content-Type: application/json

{
  "host": "127.0.0.1",
  "port": 8080,
  "auth": [
    {
      "username": "admin",
      "password": "secret123"
    },
    {
      "username": "user1",
      "password": "pass456"
    }
  ]
}
```

或使用 PUT 方法:
```http
PUT /api/config
Content-Type: application/json

{
  "host": "127.0.0.1",
  "port": 8080,
  "auth": []
}
```

**响应示例**
```json
{
  "code": 0,
  "message": "更新成功",
  "data": {
    "host": "127.0.0.1",
    "port": 8080
  }
}
```

**cURL 示例**
```bash
# POST 方式 - 启用认证
curl -X POST http://localhost:5000/api/config/create \
  -H "Content-Type: application/json" \
  -d '{"host":"127.0.0.1","port":8080,"auth":[{"username":"admin","password":"secret123"}]}'

# PUT 方式 - 禁用认证
curl -u admin:secret123 -X PUT http://localhost:5000/api/config \
  -H "Content-Type: application/json" \
  -d '{"host":"127.0.0.1","port":8080,"auth":[]}'
```

**参数说明**
- `host`: 服务地址 (必填)
- `port`: 服务端口,范围 1-65535 (必填)
- `auth`: 认证用户数组 (可选,空数组表示禁用认证)
  - `username`: 用户名 (必填,非空)
  - `password`: 密码 (必填,非空)

---

### 3. 查询单个配置项

**请求**
```http
GET /api/config/{key}
```

**支持的 key**
- `host`: 服务地址
- `port`: 服务端口
- `auth`: 认证用户数组

**响应示例**
```json
{
  "code": 0,
  "message": "查询成功",
  "data": {
    "key": "host",
    "value": "0.0.0.0"
  }
}
```

**cURL 示例**
```bash
# 查询 host
curl -u admin:secret123 http://localhost:5000/api/config/host

# 查询 port
curl -u admin:secret123 http://localhost:5000/api/config/port

# 查询 auth
curl -u admin:secret123 http://localhost:5000/api/config/auth
```

---

### 4. 修改单个配置项

**请求**
```http
PUT /api/config/{key}
Content-Type: application/json

{
  "value": "新值"
}
```

**响应示例**
```json
{
  "code": 0,
  "message": "更新成功",
  "data": {
    "key": "port",
    "value": 8080
  }
}
```

**cURL 示例**
```bash
# 修改 host
curl -u admin:secret123 -X PUT http://localhost:5000/api/config/host \
  -H "Content-Type: application/json" \
  -d '{"value":"127.0.0.1"}'

# 修改 port
curl -u admin:secret123 -X PUT http://localhost:5000/api/config/port \
  -H "Content-Type: application/json" \
  -d '{"value":8080}'

# 修改 auth (启用认证)
curl -X PUT http://localhost:5000/api/config/auth \
  -H "Content-Type: application/json" \
  -d '{"value":[{"username":"admin","password":"newpass"}]}'

# 修改 auth (禁用认证)
curl -u admin:secret123 -X PUT http://localhost:5000/api/config/auth \
  -H "Content-Type: application/json" \
  -d '{"value":[]}'
```

---

### 5. 删除配置(恢复默认值)

**请求**
```http
DELETE /api/config
```

**响应示例**
```json
{
  "code": 0,
  "message": "已恢复默认配置",
  "data": {
    "host": "0.0.0.0",
    "port": 5000
  }
}
```

**cURL 示例**
```bash
curl -u admin:secret123 -X DELETE http://localhost:5000/api/config
```

**说明**: 删除配置会将 `config.json` 恢复为默认值(包括清空 auth 数组)

---

### 6. 健康检查

**请求**
```http
GET /health
```

**响应示例**
```json
{
  "status": "ok"
}
```

**cURL 示例**
```bash
curl http://localhost:5000/health
```

---

## 错误码说明

| HTTP 状态码 | code | 说明 |
|------------|------|------|
| 200 | 0 | 成功 |
| 400 | 400 | 请求参数错误 |
| 401 | 401 | 未授权(认证失败) |
| 405 | 405 | 不支持的请求方法 |
| 415 | 415 | 不支持的媒体类型 |
| 500 | 500 | 服务器内部错误 |

**错误响应示例**
```json
// 参数错误
{
  "code": 400,
  "message": "port 必须在 1-65535 之间"
}

// 认证失败
{
  "code": 401,
  "message": "需要认证"
}
```

---

## 使用场景

### 场景1: 启用认证保护
```bash
# 1. 设置认证用户
curl -X POST http://localhost:5000/api/config/create \
  -H "Content-Type: application/json" \
  -d '{"host":"0.0.0.0","port":5000,"auth":[{"username":"admin","password":"secret123"}]}'

# 2. 之后所有请求都需要认证
curl -u admin:secret123 http://localhost:5000/api/config
```

### 场景2: 禁用认证
```bash
# 将 auth 设置为空数组
curl -u admin:secret123 -X PUT http://localhost:5000/api/config/auth \
  -H "Content-Type: application/json" \
  -d '{"value":[]}'

# 之后可以无认证访问
curl http://localhost:5000/api/config
```

### 场景3: 修改端口
```bash
# 仅修改端口
curl -u admin:secret123 -X PUT http://localhost:5000/api/config/port \
  -H "Content-Type: application/json" \
  -d '{"value":9000}'
```

### 场景4: 添加多个用户
```bash
# 设置多个认证用户
curl -u admin:secret123 -X PUT http://localhost:5000/api/config/auth \
  -H "Content-Type: application/json" \
  -d '{"value":[{"username":"admin","password":"pass1"},{"username":"user1","password":"pass2"}]}'
```

### 场景5: 重置配置
```bash
# 恢复默认配置(会清空认证)
curl -u admin:secret123 -X DELETE http://localhost:5000/api/config
```

---

## 配置文件说明

**文件位置**: `./config.json`

**默认内容**
```json
{
  "host": "0.0.0.0",
  "port": 5000,
  "auth": []
}
```

**字段说明**
- `host`: HTTP 服务监听地址
  - `0.0.0.0`: 监听所有网卡(默认)
  - `127.0.0.1`: 仅本地访问
  - `192.168.x.x`: 指定局域网 IP

- `port`: HTTP 服务端口
  - 范围: 1-65535
  - 默认: 5000
  - 注意避免与其他服务冲突

- `auth`: 认证用户列表
  - 空数组: 禁用认证(默认)
  - 非空: 启用 HTTP Basic 认证
  - 每个用户包含 `username` 和 `password`
  - 支持多个用户账号

---

## 注意事项

1. **配置持久化**: 所有修改会立即保存到 `config.json`
2. **服务重启**: 修改 `host` 和 `port` 后需要重启服务才能生效
3. **认证实时生效**: 修改 `auth` 配置立即生效,无需重启
4. **端口冲突**: 确保端口未被占用
5. **权限问题**: 确保程序有读写配置文件的权限
6. **密码安全**:
   - 密码以明文存储在配置文件中
   - 建议使用强密码
   - 生产环境建议使用 HTTPS
7. **CORS**: API 已启用 CORS,支持跨域访问
8. **健康检查**: `/health` 接口不需要认证
