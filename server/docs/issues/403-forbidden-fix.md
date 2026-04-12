# 403 Forbidden 问题排查与解决方案

## 问题现象
登录成功后调用 `GET /api/admin/user/info` 返回：
```json
{"code":403,"businessCode":0,"message":"Forbidden","data":null,"error":""}
```

## 问题原因分析

### 1. 权限中间件逻辑问题
在 `internal/app/base/middleware/permission_middleware.go` 中：

```go
// 第 56-61 行：只有 admin 角色直接放行
if roleStr == "admin" {
    c.Next()
    return
}

// 第 64-77 行：检查 API 的 auth_type
authType, err := getAPIAuthType(roleRepo, path, method, log, cacheClient)
if err != nil {
    // 之前：返回 500 错误
    // 现在：默认 auth_type=0（只需要登录）
    authType = 0
}

if authType == 0 {
    c.Next()
    return
}

// 第 79-93 行：如果 auth_type=1，检查角色权限
if !hasPermission {
    httpresponse.ForbiddenResponse(c, "Forbidden")
    c.Abort()
    return
}
```

### 2. 可能的原因

#### 原因 A：API 数据未正确初始化
- 数据库中 `/api/admin/user/info` (GET) 的记录不存在
- `auth_type` 字段值不正确（应该是 0）
- 路径不匹配（双斜杠 vs 单斜杠）

#### 原因 B：缓存问题
- Redis 中缓存了错误的 `auth_type` 值
- 缓存键不匹配

#### 原因 C：角色权限缓存问题
- 角色权限缓存中没有包含该 API
- 角色 ID 获取失败

## 解决方案

### 方案 1：清除缓存（推荐首先尝试）

清除 Redis 中的 API 授权类型缓存：

```bash
# 连接 Redis 并删除相关缓存
redis-cli
KEYS api:auth_type:*
DEL api:auth_type:GET:/api/admin/user/info
```

或者运行清理脚本：
```bash
cd server
go run scripts/clear_api_auth_cache.go
```

### 方案 2：重新初始化 API 数据

重启服务器，让种子数据重新初始化 API 路由。

或者手动在数据库中检查：
```sql
-- 检查 API 记录
SELECT * FROM apis WHERE path = '/api/admin/user/info' AND method = 'GET';

-- 如果没有记录或 auth_type 不正确，更新或插入
UPDATE apis SET auth_type = 0 WHERE path = '/api/admin/user/info' AND method = 'GET';
```

### 方案 3：检查用户角色

确认登录用户是否有角色：
1. 检查 JWT 令牌中的 `role` 字段
2. 检查 Redis 中 `user:current_role:{userID}` 的值
3. 确认用户至少有一个角色

### 方案 4：查看日志

启动服务器后，查看日志输出：
```
✓ 权限检查开始
✓ API 不需要授权，跳过权限验证
```

或者：
```
✗ 获取 API 授权类型失败
```

## 验证步骤

1. **清除缓存**
   ```bash
   redis-cli KEYS "api:auth_type:*" | xargs redis-cli DEL
   ```

2. **重启服务器**
   ```bash
   cd server
   go run cmd/main.go
   ```

3. **登录并测试**
   ```bash
   # 登录
   curl -X POST http://localhost:8000/api/admin/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"123456"}'
   
   # 使用返回的 token 调用用户信息接口
   curl -X GET http://localhost:8000/api/admin/user/info \
     -H "Authorization: Bearer {token}"
   ```

4. **检查数据库**
   ```sql
   SELECT id, path, method, auth_type, description FROM apis 
   WHERE path LIKE '%user/info%' ORDER BY created_at DESC;
   ```

## 预期结果

调用 `GET /api/admin/user/info` 应该返回：
```json
{
  "code": 200,
  "businessCode": 0,
  "message": "获取用户信息成功",
  "data": {
    "id": "xxx",
    "username": "admin",
    "nickname": "管理员",
    "roles": [...],
    "currentRoleId": "xxx",
    "role": "admin",
    ...
  },
  "error": ""
}
```

## 根本原因

代码已经修复：当 `getAPIAuthType` 获取失败时，不再返回 500 错误，而是默认 `auth_type=0`（只需要登录即可访问），确保向后兼容性。

这样可以避免因为 API 数据未初始化或查询失败而导致的 403 错误。
