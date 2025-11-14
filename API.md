# HTTP API 接口文档

## 概述

微信机器人提供 RESTful API 用于管理配置文件,支持配置的查询、创建、修改、删除操作。

**基础信息**
- 默认地址: `http://0.0.0.0:5000`
- 配置文件: `config.json`
- 响应格式: JSON

## 响应格式

所有接口返回统一格式:

```json
{
  "code": 0,              // 0=成功, 其他=失败
  "message": "操作描述",
  "data": {}              // 响应数据(可选)
}
```

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
    "port": 5000
  }
}
```

**cURL 示例**
```bash
curl http://localhost:5000/api/config
```

---

### 2. 创建/更新完整配置

**请求**
```http
POST /api/config/create
Content-Type: application/json

{
  "host": "127.0.0.1",
  "port": 8080
}
```

或使用 PUT 方法:
```http
PUT /api/config
Content-Type: application/json

{
  "host": "127.0.0.1",
  "port": 8080
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
# POST 方式
curl -X POST http://localhost:5000/api/config/create \
  -H "Content-Type: application/json" \
  -d '{"host":"127.0.0.1","port":8080}'

# PUT 方式
curl -X PUT http://localhost:5000/api/config \
  -H "Content-Type: application/json" \
  -d '{"host":"127.0.0.1","port":8080}'
```

**参数说明**
- `host`: 服务地址 (必填)
- `port`: 服务端口,范围 1-65535 (必填)

---

### 3. 查询单个配置项

**请求**
```http
GET /api/config/{key}
```

**支持的 key**
- `host`: 服务地址
- `port`: 服务端口

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
curl http://localhost:5000/api/config/host

# 查询 port
curl http://localhost:5000/api/config/port
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
curl -X PUT http://localhost:5000/api/config/host \
  -H "Content-Type: application/json" \
  -d '{"value":"127.0.0.1"}'

# 修改 port
curl -X PUT http://localhost:5000/api/config/port \
  -H "Content-Type: application/json" \
  -d '{"value":8080}'
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
curl -X DELETE http://localhost:5000/api/config
```

**说明**: 删除配置会将 `config.json` 恢复为默认值

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
| 405 | 405 | 不支持的请求方法 |
| 500 | 500 | 服务器内部错误 |

**错误响应示例**
```json
{
  "code": 400,
  "message": "port 必须在 1-65535 之间"
}
```

---

## 使用场景

### 场景1: 初次部署
```bash
# 1. 创建配置
curl -X POST http://localhost:5000/api/config/create \
  -H "Content-Type: application/json" \
  -d '{"host":"0.0.0.0","port":5000}'

# 2. 验证配置
curl http://localhost:5000/api/config
```

### 场景2: 修改端口
```bash
# 仅修改端口
curl -X PUT http://localhost:5000/api/config/port \
  -H "Content-Type: application/json" \
  -d '{"value":9000}'
```

### 场景3: 重置配置
```bash
# 恢复默认配置
curl -X DELETE http://localhost:5000/api/config
```

---

## 配置文件说明

**文件位置**: `./config.json`

**默认内容**
```json
{
  "host": "0.0.0.0",
  "port": 5000
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

---

## 注意事项

1. **配置持久化**: 所有修改会立即保存到 `config.json`
2. **服务重启**: 修改配置后需要重启服务才能生效
3. **端口冲突**: 确保端口未被占用
4. **权限问题**: 确保程序有读写配置文件的权限
5. **CORS**: API 已启用 CORS,支持跨域访问
