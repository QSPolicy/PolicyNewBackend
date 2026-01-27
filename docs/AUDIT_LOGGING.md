# 审计日志中间件使用指南

本文档介绍如何使用审计日志中间件来记录所有API请求。

---

## 📋 概述

审计日志中间件（AuditMiddleware）用于记录所有API请求的详细信息，包括：
- 请求时间戳
- 请求方法和路径
- 客户端IP地址
- 用户信息（如果已认证）
- 响应状态码
- 响应时间
- User-Agent
- 错误信息（如果有）

---

## 🚀 快速开始

### 1. 基本使用

在 `router/router.go` 中已经默认启用了审计日志：

```go
auditConfig := &custommiddleware.AuditConfig{
    EnableJSON:   false,
    LogLevel:     "info",
    LogFile:      "",
    ExcludePaths: []string{"/api/auth/login", "/api/auth/register"},
}
api.Use(custommiddleware.AuditMiddleware(auditConfig))
```

### 2. 查看日志

启动服务器后，所有API请求都会自动记录到控制台：

```bash
[2024-01-27T10:30:00+08:00] [info] GET /api/users/me => 200 (15ms)
[2024-01-27T10:30:05+08:00] [info] POST /api/teams [user:1@testuser] => 200 (45ms)
[2024-01-27T10:30:10+08:00] [info] GET /api/intelligences?page=1&pageSize=10 [user:1@testuser] => 200 (8ms)
```

---

## ⚙️ 配置选项

### AuditConfig 配置结构

```go
type AuditConfig struct {
    EnableJSON     bool
    LogLevel       string
    LogFile        string
    ExcludePaths   []string
    IncludePaths   []string
}
```

### 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| **EnableJSON** | bool | false | 是否以JSON格式输出日志 |
| **LogLevel** | string | "info" | 日志级别（info, warning, error, debug） |
| **LogFile** | string | "" | 日志文件路径（留空则只输出到控制台） |
| **ExcludePaths** | []string | [] | 排除的路径列表（不记录这些路径的请求） |
| **IncludePaths** | []string | [] | 包含的路径列表（只记录这些路径的请求，留空则记录所有） |

---

## 🎯 配置示例

### 示例1：默认配置（文本格式）

```go
auditConfig := &custommiddleware.AuditConfig{
    EnableJSON:   false,
    LogLevel:     "info",
    LogFile:      "",
    ExcludePaths: []string{"/api/auth/login", "/api/auth/register"},
}
```

**输出示例**：
```
[2024-01-27T10:30:00+08:00] [info] GET /api/users/me [user:1@testuser] => 200 (15ms)
```

### 示例2：JSON格式输出

```go
auditConfig := &custommiddleware.AuditConfig{
    EnableJSON:   true,
    LogLevel:     "info",
    LogFile:      "",
    ExcludePaths: []string{},
}
```

**输出示例**：
```json
{
    "timestamp": "2024-01-27T10:30:00+08:00",
    "level": "info",
    "remote_ip": "127.0.0.1",
    "method": "GET",
    "path": "/api/users/me",
    "query": "",
    "user_id": 1,
    "username": "testuser",
    "status_code": 200,
    "response_time_ms": 15,
    "user_agent": "Mozilla/5.0..."
}
```

### 示例3：输出到文件

```go
auditConfig := &custommiddleware.AuditConfig{
    EnableJSON:   false,
    LogLevel:     "info",
    LogFile:      "./logs/audit.log",
    ExcludePaths: []string{},
}
```

**效果**：日志同时输出到控制台和文件 `./logs/audit.log`

### 示例4：只记录特定路径

```go
auditConfig := &custommiddleware.AuditConfig{
    EnableJSON:   false,
    LogLevel:     "info",
    LogFile:      "",
    IncludePaths: []string{"/api/users/*", "/api/teams/*"},
}
```

**效果**：只记录 `/api/users/` 和 `/api/teams/` 开头的请求

### 示例5：排除特定路径

```go
auditConfig := &custommiddleware.AuditConfig{
    EnableJSON:   false,
    LogLevel:     "info",
    LogFile:      "",
    ExcludePaths: []string{
        "/api/auth/login",
        "/api/auth/register",
        "/api/health",
        "/api/search/*",
    },
}
```

**效果**：不记录登录、注册、健康检查和搜索相关的请求

---

## 📊 日志格式

### 文本格式

```
[时间戳] [级别] 方法 路径?查询参数 [user:用户ID@用户名] => 状态码 (响应时间ms) Error: 错误信息
```

**示例**：
```
[2024-01-27T10:30:00+08:00] [info] GET /api/users/me [user:1@testuser] => 200 (15ms)
[2024-01-27T10:30:05+08:00] [info] POST /api/teams [user:1@testuser] => 200 (45ms)
[2024-01-27T10:30:10+08:00] [error] GET /api/intelligences/999 [user:1@testuser] => 404 (5ms) Error: Intelligence not found
```

### JSON格式

```json
{
    "timestamp": "2024-01-27T10:30:00+08:00",
    "level": "info",
    "remote_ip": "127.0.0.1",
    "method": "GET",
    "path": "/api/users/me",
    "query": "",
    "user_id": 1,
    "username": "testuser",
    "status_code": 200,
    "response_time_ms": 15,
    "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36...",
    "error_message": ""
}
```

---

## 🎨 日志字段说明

| 字段 | 说明 | 示例 |
|------|------|------|
| **timestamp** | 请求时间戳 | 2024-01-27T10:30:00+08:00 |
| **level** | 日志级别 | info, warning, error, debug |
| **remote_ip** | 客户端IP地址 | 127.0.0.1 |
| **method** | HTTP方法 | GET, POST, PUT, DELETE |
| **path** | 请求路径 | /api/users/me |
| **query** | 查询参数 | ?page=1&pageSize=10 |
| **user_id** | 用户ID（已认证） | 1 |
| **username** | 用户名（已认证） | testuser |
| **status_code** | HTTP状态码 | 200, 404, 500 |
| **response_time_ms** | 响应时间（毫秒） | 15 |
| **user_agent** | 用户代理 | Mozilla/5.0... |
| **error_message** | 错误信息（如果有） | Intelligence not found |

---

## 🔧 高级功能

### 1. 获取当前用户信息

在任何handler中，可以使用辅助函数获取当前用户信息：

```go
import custommiddleware "policy-backend/middleware"

func MyHandler(c echo.Context) error {
    // 获取当前用户对象
    user := custommiddleware.GetCurrentUser(c)
    if user != nil {
        fmt.Printf("当前用户: %d - %s\n", user.ID, user.Username)
    }

    // 只获取用户ID
    userID := custommiddleware.GetCurrentUserID(c)
    fmt.Printf("当前用户ID: %d\n", userID)

    // 只获取用户名
    username := custommiddleware.GetCurrentUsername(c)
    fmt.Printf("当前用户名: %s\n", username)

    return nil
}
```

### 2. 手动记录日志

可以在代码中手动记录日志：

```go
import custommiddleware "policy-backend/middleware"

// 记录信息日志
custommiddleware.LogInfo("用户 %s 登录成功", username)

// 记录警告日志
custommiddleware.LogWarning("用户 %s 尝试访问未授权资源", username)

// 记录错误日志
custommiddleware.LogError("数据库查询失败: %v", err)

// 记录调试日志
custommiddleware.LogDebug("调试信息: %v", data)
```

**输出示例**：
```
[2024-01-27T10:30:00+08:00] [INFO] 用户 testuser 登录成功
[2024-01-27T10:30:05+08:00] [WARNING] 用户 testuser 尝试访问未授权资源
[2024-01-27T10:30:10+08:00] [ERROR] 数据库查询失败: record not found
[2024-01-27T10:30:15+08:00] [DEBUG] 调试信息: {"key": "value"}
```

---

## 📁 日志文件管理

### 1. 输出到文件

```go
auditConfig := &custommiddleware.AuditConfig{
    LogFile: "./logs/audit.log",
}
```

### 2. 日志轮转

建议使用日志轮转工具来管理日志文件大小：

**Linux/Mac**：使用 `logrotate`
```bash
# /etc/logrotate.d/policy-backend
/path/to/policy-backend/logs/audit.log {
    daily
    rotate 30
    compress
    delaycompress
    notifempty
    create 0644 root root
}
```

**Windows**：使用 `logrotatewin` 或类似工具

### 3. 日志分析

可以使用以下工具分析日志：

**文本格式**：
```bash
# 统计请求次数
cat logs/audit.log | wc -l

# 统计各状态码数量
cat logs/audit.log | grep -o '=> [0-9]*' | sort | uniq -c

# 查找错误请求
cat logs/audit.log | grep "Error"

# 按响应时间排序
cat logs/audit.log | sort -t'(' -k2 -n
```

**JSON格式**：使用 `jq`
```bash
# 统计各状态码数量
cat logs/audit.log | jq -r '.status_code' | sort | uniq -c

# 查找响应时间超过100ms的请求
cat logs/audit.log | jq 'select(.response_time_ms > 100)'

# 统计各路径的平均响应时间
cat logs/audit.log | jq -s 'group_by(.path) | map({path: .[0].path, avg_time: (map(.response_time_ms) | add / length)})'
```

---

## 🐛 故障排除

### 问题1：没有看到日志输出

**解决方案**：
- 检查 `LogFile` 路径是否正确
- 确保 `ExcludePaths` 没有排除你要记录的路径
- 检查 `IncludePaths` 是否设置了正确的路径

### 问题2：日志文件权限错误

**解决方案**：
```bash
# 确保日志目录存在且有写入权限
mkdir -p logs
chmod 755 logs
chmod 644 logs/audit.log
```

### 问题3：日志文件太大

**解决方案**：
- 配置日志轮转
- 定期清理旧日志
- 使用更详细的 `ExcludePaths` 排除不需要的请求

### 问题4：JSON格式输出不正确

**解决方案**：
- 确保 `EnableJSON` 设置为 `true`
- 检查日志内容是否包含特殊字符
- 使用 `jq` 验证JSON格式

---

## 📝 最佳实践

### 1. 性能考虑

- ✅ 生产环境建议使用JSON格式，便于日志分析
- ✅ 排除健康检查和心跳接口
- ✅ 定期清理旧日志
- ✅ 考虑使用异步日志写入

### 2. 安全考虑

- ✅ 不要记录敏感信息（密码、token等）
- ✅ 日志文件设置适当的权限（0644）
- ✅ 考虑加密存储敏感日志
- ✅ 定期审计日志访问

### 3. 可维护性

- ✅ 使用一致的日志格式
- ✅ 为不同环境配置不同的日志级别
- ✅ 文档化日志字段含义
- ✅ 定期审查日志配置

---

## 🎉 总结

审计日志中间件提供了：

✅ **完整的请求记录** - 记录所有必要的请求信息
✅ **灵活的配置** - 支持文本/JSON格式、文件输出、路径过滤
✅ **易用的API** - 提供辅助函数获取用户信息和手动记录日志
✅ **高性能** - 轻量级实现，对性能影响小
✅ **生产就绪** - 适合生产环境使用

现在你可以：
- 📊 监控API使用情况
- 🔍 排查问题和调试
- 📈 分析性能瓶颈
- 🔐 安全审计和合规

祝使用愉快！ 🚀
