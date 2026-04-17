***

name: 全栈开发
description: 进行全栈开发时触发，包括新功能开发、功能修改、Bug 修复、代码重构等场景，提供从需求分析到提交部署的完整 7 步流程指导，确保前后端规范统一
------------------------------------------------------------------------------------

## 核心原则：前后端交互规范 ⭐

### 1. JSON 序列化命名规范

**统一使用 camelCase（小驼峰命名）**

- ✅ **正确**: `userId`, `createdAt`, `currentRoleId`
- ❌ **错误**: `user_id` (snake\_case), `UserID` (PascalCase)

**原因**:

- 前端 TypeScript/JavaScript 使用 camelCase
- 保持前后端命名一致性
- 符合行业标准

### 2. ID 字段处理规范 ⭐⭐⭐

**采用 `json:",string"` 标签方案，全层级统一使用 `int64` 类型**

- ✅ **正确**: `ID int64 `json:"id,string"`` → JSON 序列化为 `"id": "696243340630298624"`
- ❌ **错误**: `ID int64 `json:"id"`` → JSON 序列化为 `"id": 696243340630298624`（精度丢失）

**原因**: JavaScript Number 精度限制（±2^53），Go int64 会丢失精度

**单个 ID 字段实现方案**:

```Go
// Domain Entity (领域层)
type User struct {
    ID           int64  // ✅ 使用 int64，与数据库一致
    DepartmentID int64  // ✅ 使用 int64
}

// Response DTO (接口层)
type UserResponse struct {
    ID           int64  `json:"id,string"`           // ✅ 添加 ,string 标签，自动序列化为字符串
    DepartmentID int64  `json:"departmentId,string"` // ✅ 添加 ,string 标签
}

// Request DTO (接口层)
type UpdateUserRequest struct {
    ID           int64  `json:"id,string"`           // ✅ 添加 ,string 标签，自动解析字符串为 int64
    DepartmentID int64  `json:"departmentId,string"` // ✅ 添加 ,string 标签
}
```

**ID 数组字段实现方案**（Go 的 `json:",string"` 标签不支持数组）:

```Go
// Request DTO (ID 数组) - 使用 []string
type AssignUserToRoleRequest struct {
    RoleID  int64    `json:"roleId" binding:"required"`
    UserIDs []string `json:"userIds" binding:"required,min=1"`  // ✅ 使用 []string
}

// Response DTO (ID 数组) - 使用 []int64
type APIResponse struct {
    RoleIds    []int64 `json:"roleIds"`    // ✅ 使用 []int64
    MenuIds    []int64 `json:"menuIds"`    // ✅ 使用 []int64
    BtnPermIds []int64 `json:"btnPermIds"` // ✅ 使用 []int64
}

// Domain Entity (ID 数组) - 使用 []int64
type User struct {
    RoleIds []int64     // ✅ 使用 []int64
    Roles   []*Role     // ✅ 关联对象列表
}
```

**API Handler 层转换（[]string → []int64）**:

```Go
func (c *RoleController) AssignUserToRole(ctx *gin.Context) {
    var req AssignUserToRoleRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        return
    }

    // ✅ 将 []string 转换为 []int64
    userIDInts := make([]int64, len(req.UserIDs))
    for i, idStr := range req.UserIDs {
        id, err := strconv.ParseInt(idStr, 10, 64)
        if err != nil {
            baseRes.FailWithMessage("用户 ID 格式错误", ctx)
            return
        }
        userIDInts[i] = id
    }

    // 调用 Service（使用 []int64）
    err := c.service.AssignUserToRole(req.RoleID, userIDInts)
    // ...
}
```

**ID 字段类型使用规则**:

| 层级 | 单个 ID | ID 数组 | 说明 |
|------|--------|--------|------|
| **Request DTO** | `int64` + `json:",string"` | `[]string` | Go 的 `json:",string"` 不支持数组 |
| **Response DTO** | `int64` + `json:",string"` | `[]int64` | 自动序列化/反序列化 |
| **Domain Entity** | `int64` | `[]int64` | 与数据库一致，无需转换 |

**优势**:

1. **类型安全**：全层级使用 int64，无需类型转换代码
2. **代码简洁**：Repository 层直接赋值，不需要 string ↔ int64 转换
3. **自动转换**：Go json 标签自动处理 int64 ↔ string 序列化
4. **性能更好**：无类型转换开销

#### 2.1 ID 数组字段处理规范 ⭐⭐⭐

**Go 的 `json:",string"` 标签不支持数组类型，需要使用 `[]string`**

**Request DTO (ID 数组)**:
```go
// ✅ 正确：使用 []string 接收前端传来的字符串数组
type AssignUserToRoleRequest struct {
    RoleID  int64    `json:"roleId" binding:"required"`
    UserIDs []string `json:"userIds" binding:"required,min=1"`  // ✅ 使用 []string
}

// ✅ 正确：批量删除操作
type BatchDeleteRequest struct {
    IDs []int64 `json:"ids" binding:"required"`  // ✅ 单个 ID 用 json:",string"，数组用 []int64
}
```

**Response DTO (ID 数组)**:
```go
// ✅ 正确：使用 []int64（不需要 json:",string" 标签）
type APIResponse struct {
    RoleIds    []int64 `json:"roleIds"`    // ✅ 使用 []int64
    MenuIds    []int64 `json:"menuIds"`    // ✅ 使用 []int64
    BtnPermIds []int64 `json:"btnPermIds"` // ✅ 使用 []int64
}
```

**Domain Entity (ID 数组)**:
```go
// ✅ 正确：领域层使用 []int64
type User struct {
    RoleIds []int64     // ✅ 使用 []int64
    Roles   []*Role     // ✅ 关联对象列表
}
```

**API Handler 层转换（[]string → []int64）**:
```go
func (c *RoleController) AssignUserToRole(ctx *gin.Context) {
    var req AssignUserToRoleRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        return
    }

    // ✅ 将 []string 转换为 []int64
    userIDInts := make([]int64, len(req.UserIDs))
    for i, idStr := range req.UserIDs {
        id, err := strconv.ParseInt(idStr, 10, 64)
        if err != nil {
            baseRes.FailWithMessage("用户 ID 格式错误", ctx)
            return
        }
        userIDInts[i] = id
    }

    // 调用 Service（使用 []int64）
    err := c.service.AssignUserToRole(req.RoleID, userIDInts)
    // ...
}
```

**Domain Service 层**:
```go
// ✅ 直接接收 []int64，不需要转换
func (s *RoleService) AssignUserToRole(roleID int64, userIDs []int64) error {
    for _, userID := range userIDs {
        // 业务逻辑
    }
    return nil
}
```

**ID 数组类型使用规则**:

| 层级 | ID 数组类型 | 说明 |
|------|-----------|------|
| **Request DTO** | `[]string` | 前端传来 JSON 字符串数组，Go 的 `json:",string"` 不支持数组 |
| **Response DTO** | `[]int64` | 直接返回 int64 数组，Go 会自动序列化 |
| **Domain Entity** | `[]int64` | 与数据库一致，无需转换 |
| **API Handler** | 手动转换 `[]string` → `[]int64` | 在边界层完成类型转换 |

**为什么这样设计？**

1. **Go 限制**：`json:",string"` 标签只支持标量类型，不支持数组
2. **前端传来**：JSON 格式为 `{"userIds": ["123", "456"]}`（字符串数组）
3. **边界转换**：在 API Handler 层完成类型转换，保证内部使用统一的 `int64`
4. **类型安全**：Domain Entity 使用 `[]int64`，编译期类型检查

### 3. CRUD 接口命名规范 ⭐⭐⭐

**统一 CRUD 基本操作的 HTTP 方法和命名规则**

| 操作       | HTTP 方法 | 路由模式            | 命名模式              | 示例                                          |
| -------- | ------- | --------------- | ----------------- | ------------------------------------------- |
| **创建**   | POST    | `/{entity}`     | `Create{Entity}`  | `CreateUser`, `CreateRole`, `CreateMenu`    |
| **更新**   | PUT     | `/{entity}`     | `Update{Entity}`  | `UpdateUser`, `UpdateRole`, `UpdateMenu`    |
| **删除**   | DELETE  | `/{entity}/:id` | `Delete{Entity}`  | `DeleteUser`, `DeleteRole`, `DeleteMenu`    |
| **查询单条** | GET     | `/{entity}/:id` | `Get{Entity}ByID` | `GetUserByID`, `GetRoleByID`, `GetMenuByID` |

**示例代码**：

```Go
// 创建用户
createUser({ userName: 'admin', email: 'admin@example.com' })
// POST /api/admin/user

// 更新用户
updateUser({ id: '123', userName: 'admin', email: 'admin@example.com' })
// PUT /api/admin/user

// 删除用户
deleteUser('123')
// DELETE /api/admin/user/123

// 查询单条
getUserByID('123')
// GET /api/admin/user/123
```

**注意事项**：

- ✅ 统一使用 `Create` 前缀，不使用 `Add`
- ✅ 统一使用 `Update` 前缀
- ✅ 统一使用 `Delete` 前缀
- ✅ 统一使用 `Get{Entity}ByID` 查询单条，不使用 `Detail`

### 4. API 接口种子数据管理 ⭐⭐⭐

**开发新增 API 接口后，必须在种子数据中注册**

**位置**: `internal/domain/base/seed/apis_seed.go`

**步骤**：

1. **在** **`getAPIRoutes()`** **方法中添加 API 路由定义**

```Go
func (as *APISeed) getAPIRoutes() []*entity.API {
    apis := []*entity.API{
        // ==================== 用户管理 ====================
        {
            Path:         "/api/admin/user",
            Method:       "POST",
            Group:        "用户管理",
            AuthRequired: true,
            AuthType:     1,
            Description:  "创建用户",
            Status:       1,
        },
        {
            Path:         "/api/admin/user/:id",
            Method:       "PUT",
            Group:        "用户管理",
            AuthRequired: true,
            AuthType:     1,
            Description:  "更新用户",
            Status:       1,
        },
        // ... 其他 API 路由
    }
    return apis
}
```

1. **API 路由字段说明**

| 字段             | 说明      | 示例                             |
| -------------- | ------- | ------------------------------ |
| `Path`         | API 路径  | `/api/admin/user`              |
| `Method`       | HTTP 方法 | `POST`, `GET`, `PUT`, `DELETE` |
| `Group`        | 所属模块    | `用户管理`, `角色管理`                 |
| `AuthRequired` | 是否需要认证  | `true`, `false`                |
| `AuthType`     | 认证类型    | `0`-普通认证，`1`-管理员认证             |
| `Description`  | 接口描述    | `创建用户`, `更新用户`                 |
| `Status`       | 状态      | `1`-启用，`0`-禁用                  |

1. **种子数据初始化逻辑**

- ✅ **幂等性**：自动检查是否存在，不存在则创建，存在则更新
- ✅ **增量写入**：不会删除已有的 API 路由
- ✅ **自动同步**：应用启动时自动执行

1. **完整示例**

```Go
// 在 getAPIRoutes() 中添加新的 API 路由
{
    Path:         "/api/admin/user",
    Method:       "POST",
    Group:        "用户管理",
    AuthRequired: true,
    AuthType:     1,      // 需要管理员权限
    Description:  "创建用户",
    Status:       1,      // 启用状态
},
```

**注意事项**：

- ✅ 所有新增的 API 接口都必须在此注册
- ✅ 路径必须与 routes.go 中的定义一致
- ✅ 根据接口权限设置正确的 `AuthType`
- ✅ 使用中文描述接口功能

### 5. 列表接口统一设计 ⭐⭐⭐

**统一使用** **`GET /{entity}`** **作为列表接口，通过参数控制行为**

**参数灵活组合**：

- 带 `page`, `pageSize` → 返回 PageResult（分页数据）
- 不带分页参数 → 返回数组（所有匹配结果）
- 支持搜索参数（如 `name`, `status`, `keyword` 等）

**示例**：

```Go
// 分页列表
getRoleList({ page: 1, pageSize: 10, name: '管理员' })
// 返回：{ list: [...], total: 100, page: 1, pageSize: 10 }

// 搜索（不分页）
getRoleList({ name: '管理员' })
// 返回：[所有匹配的角色数组]

// 获取全部
getRoleList()
// 返回：[所有角色数组]
```

**优势**：

- ✅ 减少接口数量 - 列表和搜索使用同一个接口
- ✅ 更灵活 - 前端可以根据需要选择是否分页
- ✅ 更符合 RESTful - 资源统一，通过参数控制行为
- ✅ 向后兼容 - 现有调用不受影响

***

## 核心架构：数据流转与转换规范 ⭐⭐⭐

### 一、完整的数据流转路径

#### 1. 写入流程（API → Database）

```
HTTP 请求 → API Handler → Domain Service → Repository → Database
          (DTO 层)      (领域实体)    (数据模型)   (数据库)
          (int64)       (int64)       (int64)      (int64)
```

**ID 数组字段流转**：

```
HTTP 请求：{"userIds": ["123", "456"]}
   ↓
Request DTO: UserIDs []string  ← Go json:",string" 不支持数组
   ↓
API Handler: 手动转换 []string → []int64（使用 strconv.ParseInt）
   ↓
Domain Service / Entity: []int64  ← 直接使用，无需转换
   ↓
Repository: 保存到数据库（[]int64）
```

**详细步骤**:

```
┌─────────────────────────────────────────────────────────┐
│ 1. HTTP POST /api/admin/base/user                      │
│    Body: {"userName": "admin", "email": "admin@ix.com"}│
│    ID 字段：json:",string" 标签自动解析为 int64         │
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│ 2. API Handler (internal/app/base/api/user_handler.go) │
│    - ShouldBindJSON → RegisterRequest DTO              │
│    - 验证参数：binding:"required,email"                │
│    - ID 字段：int64 类型（json:",string" 标签）          │
│    - 调用 service.Register(req.Username, req.Email)    │
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│ 3. Domain Service (internal/domain/base/service/)      │
│    - 业务逻辑：密码加密、邮箱验证                       │
│    - 创建 Domain Entity（ID: int64）                    │
│    - 调用 repo.Create(user)                            │
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│ 4. Repository (internal/persistence/base/)             │
│    - ⭐ 转换：Domain Entity → Database Model            │
│    - fromDomain(): 直接赋值（无需类型转换）             │
│    - 调用 db.Create(dbModel)                           │
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│ 5. PostgreSQL 数据库                                    │
│    - INSERT INTO base_users (id, userName, email)      │
│    - 返回：雪花算法 ID (int64)                         │
└─────────────────────────────────────────────────────────┘
```

#### 2. 查询流程（Database → API）

```
PostgreSQL → Database Model → Domain Entity → API Handler → Response DTO → HTTP 响应
   (int64)      toDomain()     (int64)       手动构建      (int64+json:",string")
                                              (直接赋值)    (自动序列化为字符串)
```

**详细步骤**:

```
┌─────────────────────────────────────────────────────────┐
│ 1. HTTP GET /api/admin/base/user/info                  │
│    Header: Authorization: Bearer {token}               │
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│ 2. API Handler (internal/app/base/api/user_handler.go) │
│    - 从 token 解析 userID (int64)                        │
│    - 调用 service.GetUserInfo(userID)                  │
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│ 3. Domain Service (internal/domain/base/service/)      │
│    - 调用 repo.GetByID(userID)                         │
│    - 返回：*entity.User（Domain Entity, ID: int64）    │
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│ 4. Repository (internal/persistence/base/)             │
│    - SQL: SELECT * FROM base_users WHERE id = ?        │
│    - 返回：userModel (Database Model, ID: int64)       │
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│ 5. Repository toDomain()                               │
│    - ⭐ 直接赋值：user.ID = m.ID（无需转换）            │
│    - 返回：*entity.User（Domain Entity, ID: int64）    │
└─────────────────────────────────────────────────────────┘
                        ↓ (逐层返回)
┌─────────────────────────────────────────────────────────┐
│ 6. API Handler 构建 DTO                                 │
│    - ⭐ 手动映射：entity.User → UserResponse           │
│    - ID 字段：直接赋值 int64                           │
│    - json:",string" 标签自动序列化为字符串              │
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│ 7. HTTP 响应                                             │
│    Body: {"id": "696243340630298624", "userName": "..."}│
│    ID 字段：自动序列化为字符串格式                       │
└─────────────────────────────────────────────────────────┘
```
│ 8. HTTP Response                                        │
│    {                                                    │
│      "code": 200,                                       │
│      "data": {                                          │
│        "id": "1234567890",  ✅ string                  │
│        "userName": "admin",                             │
│        "roles": [{"id": "1", "name": "管理员"}]         │
│      }                                                  │
│    }                                                    │
└─────────────────────────────────────────────────────────┘
```

### 二、转换层次总结表

| 转换阶段                                 | 位置             | 转换方向                  | 转换方法                                                                                         | 示例代码位置                                                                                            |
| ------------------------------------ | -------------- | --------------------- | -------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------- |
| **HTTP → DTO**                       | API Handler    | JSON → Request DTO    | `ctx.ShouldBindJSON()`                                                                       | [`user_handler.go`](d:\g\ixpay-pro\server\internal\app\base\api\user_handler.go)                  |
| **DTO → 参数**                         | API Handler    | Request DTO → 简单参数    | 直接传递字段                                                                                       | [`user_handler.go`](d:\g\ixpay-pro\server\internal\app\base\api\user_handler.go)                  |
| **参数 → Domain Entity**               | Domain Service | 简单参数 → Entity         | 构造器创建                                                                                        | [`user_service.go`](d:\g\ixpay-pro\server\internal\domain\base\service\user_service.go)           |
| **Domain Entity → Database Model** ⭐ | Repository     | Entity → Model        | [`fromDomain()`](d:\g\ixpay-pro\server\internal\persistence\base\user_repository.go#L64-L90) | [`user_repository.go:64`](d:\g\ixpay-pro\server\internal\persistence\base\user_repository.go#L64) |
| **Database Model → Domain Entity** ⭐ | Repository     | Model → Entity        | [`toDomain()`](d:\g\ixpay-pro\server\internal\persistence\base\user_repository.go#L37-L61)   | [`user_repository.go:37`](d:\g\ixpay-pro\server\internal\persistence\base\user_repository.go#L37) |
| **Domain Entity → DTO**              | API Handler    | Entity → Response DTO | 手动构建                                                                                         | [`user_handler.go`](d:\g\ixpay-pro\server\internal\app\base\api\user_handler.go)                  |

### 三、关键设计原则

#### 1. Repository 层是数据转换的核心 ⭐⭐⭐

**所有** `Database Model ↔ Domain Entity` 的转换都在 Repository 层完成，**无需类型转换**：

```Go
// internal/persistence/base/user_repository.go

// ⭐ Database Model → Domain Entity（查询时使用）
func (m *userModel) toDomain() *entity.User {
    user := &entity.User{
        ID:            m.ID,            // ✅ 直接赋值（int64 → int64）
        Username:      m.Username,
        PositionID:    m.PositionID,    // ✅ 直接赋值（int64 → int64）
        DepartmentID:  m.DepartmentID,  // ✅ 直接赋值（int64 → int64）
        CreatedBy:     m.CreatedBy,     // ✅ 直接赋值（int64 → int64）
    }
    
    // 处理关联数据
    if m.Department != nil {
        user.Department = m.Department.toDomain()
    }
    if m.Position != nil {
        user.Position = m.Position.toDomain()
    }
    if len(m.Roles) > 0 {
        roles := make([]*entity.Role, len(m.Roles))
        roleIDs := make([]int64, len(m.Roles))
        for i, role := range m.Roles {
            roles[i] = role.toDomain()
            roleIDs[i] = role.ID
        }
        user.Roles = roles
        user.RoleIds = roleIDs
    }
    
    return user
}

// ⭐ Domain Entity → Database Model（写入时使用）
func fromDomain(user *entity.User) (*userModel, error) {
    return &userModel{
        SnowflakeBaseModel: database.SnowflakeBaseModel{
            ID:        user.ID,            // ✅ 直接赋值（int64 → int64）
            CreatedBy: user.CreatedBy,     // ✅ 直接赋值（int64 → int64）
            UpdatedBy: user.UpdatedBy,     // ✅ 直接赋值（int64 → int64）
        },
        Username:     user.Username,
        PositionID:   user.PositionID,     // ✅ 直接赋值（int64 → int64）
        DepartmentID: user.DepartmentID,   // ✅ 直接赋值（int64 → int64）
    }, nil
}
```

**Repository 方法中的转换示例**:

```Go
// 查询：GetByID
func (r *userRepository) GetByID(id int64) (*entity.User, error) {
    // 1️⃣ 查询数据库（直接使用 int64）
    var dbModel userModel
    result := r.db.Where("id = ?", id).First(&dbModel)
    
    // 2️⃣ Database Model → Domain Entity
    return dbModel.toDomain(), nil  // ✅ 返回 Domain Entity
}

// 写入：Create（⭐ 包含 ID 回写）
func (r *userRepository) Create(user *entity.User) error {
    // 1️⃣ Domain Entity → Database Model
    dbModel, err := fromDomain(user)
    if err != nil {
        return err
    }
    
    // 2️⃣ 保存到数据库（分两步：先保存，再回写 ID）
    if err := r.db.Create(dbModel).Error; err != nil {
        return err
    }
    
    // 3️⃣ ⭐ 将生成的 ID 回写到领域实体（重要！）
    user.ID = dbModel.ID  // ✅ 直接赋值（int64 → int64）
    return nil
}
```

**⭐⭐⭐ ID 回写最佳实践 ⭐⭐⭐**

**为什么需要 ID 回写？**

当使用雪花算法生成 ID 时，数据库 Model 在 Create 后会获得一个 int64 类型的 ID。由于领域实体使用 string 类型的 ID，必须在 Create 方法完成后将生成的 ID 回写到领域实体，以确保：

1. ✅ **领域实体完整性**：调用方可以立即使用新生成的 ID
2. ✅ **避免二次查询**：不需要再次查询数据库获取 ID
3. ✅ **事务一致性**：在同一事务内完成 ID 回写
4. ✅ **类型转换**：将 int64 转换为 string，保持前后端一致性

#### 2. Domain Service 层保持纯净

- ✅ **不做数据转换**：只处理业务逻辑
- ✅ **不接触 DTO**：接收简单参数或领域实体
- ✅ **不接触 Database Model**：通过 Repository 接口操作数据
- ✅ **直接透传 Domain Entity**

```Go
// internal/domain/base/service/user_service.go

// ✅ 正确：接收简单参数，返回 Domain Entity
func (s *UserService) Register(userName, password, email string) (*entity.User, error) {
    // 业务逻辑：密码加密
    passwordHash := encryption.GeneratePasswordHash(password)
    
    // 创建 Domain Entity
    user := &entity.User{
        Username:     userName,
        PasswordHash: passwordHash,
        Email:        email,
    }
    
    // 调用 Repository（传递 Domain Entity）
    if err := s.repo.Create(user); err != nil {
        return nil, err
    }
    
    return user, nil  // ✅ 返回 Domain Entity
}
```

#### 3. API Handler 层负责 DTO 映射

**职责**:

- ✅ 接收请求 → 绑定到 Request DTO
- ✅ 调用 Domain Service → 获取 Domain Entity
- ✅ 手动构建 Response DTO → 返回 HTTP 响应

```Go
// internal/app/base/api/user_handler.go

// ⭐ 单条查询：手动构建 Response DTO
func (h *UserHandler) GetUserInfo(ctx *gin.Context) {
    // 1️⃣ 调用 Domain Service，返回 Domain Entity
    user, err := h.service.GetUserInfo(userID.(int64))
    
    // 2️⃣ ⭐ 手动转换：Domain Entity → Response DTO
    roleInfos := make([]*response.RoleInfo, 0, len(user.Roles))
    for _, role := range user.Roles {
        roleInfo := &response.RoleInfo{
            ID:          role.ID,        // ✅ string 类型
            Name:        role.Name,
            Code:        role.Code,
        }
        roleInfos = append(roleInfos, roleInfo)
    }
    
    // 3️⃣ 构建最终响应 DTO
    response := response.UserInfoResponse{
        ID:            user.ID,           // ✅ string → string
        Username:      user.Username,
        Nickname:      user.Nickname,
        Roles:         roleInfos,
        CurrentRoleId: finalRoleID,
    }
    
    // 4️⃣ 返回 HTTP 响应
    baseRes.OkWithDetailed(response, "获取用户信息成功", ctx)
}
```

#### 4. 类型一致性策略

| 层级                | ID 类型              | 原因                      |
| ----------------- | ------------------ | ----------------------- |
| **数据库存储**         | `int64`            | 索引性能好，存储空间小，使用雪花算法      |
| **Domain Entity** | `int64`            | 与数据库一致，无需类型转换         |
| **DTO**           | `int64` + `json:",string"` | API 传输安全，json 标签自动序列化为字符串 |

**转换责任**: 无需转换，全层级统一使用 `int64`，通过 `json:",string"` 标签实现序列化

***

## 标准 7 步开发流程

### 流程步骤

1. **需求分析与 API 设计**（前后端对齐）
   - 分析需求，明确功能边界
   - 设计 API 接口（后端优先）
   - 与前端确认接口设计
2. **后端开发**（分层实现）
   - 生成后端脚手架（api/handler, application/service, domain/entity, persistence/repository）
   - 实现业务逻辑
   - 编写单元测试
   - 运行测试确保通过
   - 格式化代码
   - 生成 API 文档
3. **后端自测与验证**
   - 编译检查（go build）
   - 运行测试（go test -cover > 80%）
   - API 测试（Swagger/Postman）
   - 代码审查（设计原则 + 规范）
4. **前端开发**（基于 API 接口）
   - 确认 API 接口文档
   - 生成前端组件
   - 实现页面逻辑
   - 联调测试
5. **前后端联调**
   - 启动后端服务
   - 启动前端服务器
   - 联调测试
   - 修复问题
6. **代码审查与优化**
   - 全面审查（规范 + 设计原则）
   - 自动修复简单问题
   - 手动修复复杂问题
7. **提交与部署**
   - 最终检查
   - 格式化代码
   - 更新文档
   - 提交代码
   - 部署（如需要）

***

## 后端规范（Go + Gin + DDD 分层架构）

**核心架构说明**：数据流转与转换规范详见 [核心架构](#核心架构数据流转与转换规范) 章节。

### 0. Go 代码风格规范

#### 代码格式

- 4 空格缩进
- 每行≤120 字符
- K\&R 括号风格
- 运算符两侧加空格

#### 命名规范

- 包名：小写字母
- 文件名：小写 + 下划线
- 函数/变量：驼峰命名
- 常量：大写 + 下划线
- 结构体：PascalCase
- JSON 标签：camelCase

#### 导入顺序

标准库 → 第三方库 → 本地包（组间空行）

#### 代码结构

- 函数≤50 行
- 避免深层嵌套
- 优先提前返回
- 遵循单一职责

#### 切片处理

- ✅ `len(slice) > 0`
- ❌ `slice != nil && len(slice) > 0`
- 初始化：`make([]T, 0)` 或 `[]T{}`

#### 数据库操作

- ✅ 必须使用 GORM
- ❌ 禁止原生 SQL（migrations 除外）
- ✅ 对象条件：`Where(&Model{ID: id})`
- ❌ 字符串条件：`Where("id = ?", id)`

#### 注释规范

- 公共函数必须注释
- 使用中文注释

### 1. 目录结构规范

```
server/
├── cmd/                    # 应用程序入口
│   └── ixpay-pro/
│       └── main.go
├── configs/                # 配置文件
├── docs/                   # Swagger 文档（自动生成）
├── internal/               # 核心业务代码
│   ├── app/                # 应用层
│   │   ├── base/           # 基础管理模块
│   │   │   ├── api/        # API 处理器
│   │   │   ├── application/ # 应用服务
│   │   │   ├── middleware/ # 模块中间件
│   │   │   ├── migrations/ # 数据库迁移
│   │   │   ├── seed/       # 数据库种子
│   │   │   ├── application.go
│   │   │   ├── wire.go     # 依赖注入配置
│   │   │   └── routes.go   # 路由配置
│   ├── domain/             # 领域层
│   │   ├── base/           # 基础管理领域
│   │   │   ├── entity/     # 领域实体
│   │   │   ├── repo/       # 仓库接口
│   │   │   │   └── mock/   # Mock 实现
│   │   │   └── service/    # 领域服务
│   ├── persistence/        # 持久化层
│   │   ├── base/           # 基础管理模块仓库实现
│   │   └── common/         # 通用持久化工具
│   ├── infrastructure/     # 基础设施层
│   │   ├── persistence/    # 数据持久化
│   │   ├── transport/      # 传输层
│   │   ├── security/       # 安全相关
│   │   ├── observability/  # 可观测性
│   │   └── support/        # 支撑工具
│   ├── dto/                # 数据传输对象
│   │   ├── base/
│   │   │   ├── request/    # 请求 DTO
│   │   │   └── response/   # 响应 DTO
│   ├── utils/              # 工具函数
│   └── config/             # 配置管理
├── tests/                  # 测试目录
│   ├── unit/               # 单元测试
│   ├── integration/        # 集成测试
│   └── e2e/                # 端到端测试
├── scripts/                # 脚本目录
├── .trae/                  # 项目配置
└── go.mod                  # Go 模块文件
```

### 2. 响应格式规范

所有 API 响应必须遵循统一格式：

```json
{
  "code": 0,
  "data": {},
  "msg": "success"
}
```

**要求**:

- 使用 `baseRes.OkWithDetailed(data, msg, ctx)` 统一响应
- `msg` 字段使用中文描述
- 禁止直接返回数据

### 3. JSON 标签命名规范

**JSON 标签命名规范已在** **[核心原则 - JSON 序列化命名](#1-json-序列化命名规范)** **中定义，此处不再重复。**

### 4. 领域实体定义

```Go
// internal/domain/base/entity/user.go
package entity

// User 领域实体（无 GORM 标签）
type User struct {
    ID       string   // ✅ 使用 string 类型
    Username string
    Email    string
    RoleIds  []string // ✅ 使用 string 数组
    Password string
}

// 领域方法：检查用户是否有某个角色
func (u *User) HasRole(roleID string) bool {
    for _, rid := range u.RoleIds {
        if rid == roleID {
            return true
        }
    }
    return false
}
```

### 4.1 关联数据设计模式：同时保留 ID 列表和完整对象 ⭐⭐⭐

**设计原则**：在领域实体中同时保留关联对象的 ID 列表和完整对象列表

**示例代码**：

```Go
// internal/domain/base/entity/user.go
package entity

// User 领域实体
type User struct {
    // ... 基本字段
    
    // ⭐ 关联数据：同时保留 ID 列表和完整对象
    RoleIds     []string    // 用户关联的角色 ID 列表（轻量级，用于业务判断）
    Roles       []*Role     // 角色列表（完整对象，用于返回详细信息）
    
    DepartmentID string     // 部门 ID
    Department  *Department // 所属部门对象
    
    PositionID string       // 岗位 ID
    Position  *Position     // 岗位对象
}
```

**优势对比**：

| 设计方式          | 性能    | 灵活性   | 内存占用  | 适用场景         |
| ------------- | ----- | ----- | ----- | ------------ |
| **只保留 ID 列表** | ⭐⭐⭐⭐⭐ | ⭐⭐    | ⭐⭐⭐⭐⭐ | 只需判断关联关系     |
| **只保留完整对象**   | ⭐⭐    | ⭐⭐⭐⭐  | ⭐⭐    | 需要完整信息展示     |
| **两者都保留** ⭐   | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐   | **推荐：适应多场景** |

**使用场景**：

1. **登录场景**（使用 ID 列表）：

```Go
// 登录时只需要角色 ID 生成 JWT token
user, err := s.repo.GetByID(userID)  // 不加载 Roles
if len(user.RoleIds) > 0 {
    roleCode = user.RoleIds[0]  // ✅ 使用 RoleIds，轻量高效
}
```

1. **用户详情场景**（使用完整对象）：

```Go
// 获取用户详情，需要展示完整角色信息
user, err := s.repo.GetByID(userID, repo.UserRelationRoles)  // 加载 Roles
// ✅ 使用 Roles，返回完整角色信息给前端
roleInfos := make([]*response.RoleInfo, 0, len(user.Roles))
for _, role := range user.Roles {
    roleInfos = append(roleInfos, &response.RoleInfo{
        ID:   role.ID,
        Name: role.Name,
        Code: role.Code,
    })
}
```

1. **业务判断场景**（使用 ID 列表）：

```Go
// 权限检查：使用 RoleIds 高效判断
func (u *User) HasRole(roleID string) bool {
    for _, rid := range u.RoleIds {  // ✅ 字符串比较，高效
        if rid == roleID {
            return true
        }
    }
    return false
}
```

**Repository 层转换**：

```Go
// internal/persistence/base/user_repository.go
func (m *userModel) toDomain() *entity.User {
    user := &entity.User{
        ID:       common.ToString(m.ID),
        Username: m.Username,
        // ... 基本字段
    }
    
    // ⭐ 处理关联数据：同时填充 ID 列表和完整对象
    if len(m.Roles) > 0 {
        // 一个循环完成两个任务
        roles := make([]*entity.Role, len(m.Roles))
        roleIDs := make([]string, len(m.Roles))
        for i, role := range m.Roles {
            roles[i] = role.toDomain()           // ✅ 转换为领域实体
            roleIDs[i] = common.ToString(role.ID) // ✅ 填充 ID 列表
        }
        user.Roles = roles
        user.RoleIds = roleIDs
    }
    
    return user
}
```

**注意事项**：

1. ✅ **数据一致性**：在 `toDomain()` 方法中同时填充两个字段，确保一致性
2. ✅ **性能优化**：使用一个循环完成两个任务，避免多次遍历
3. ✅ **按需加载**：通过 Preload 控制是否加载完整对象，避免过度查询
4. ✅ **职责分离**：
   - `RoleIds`：用于业务逻辑判断、权限检查、生成 Token
   - `Roles`：用于返回完整信息、展示角色详情

**其他关联关系同样适用**：

```Go
type Department struct {
    ID         string
    Name       string
    ParentID   string        // 父部门 ID
    Parent     *Department   // 父部门对象
    LeaderID   string        // 负责人 ID
    Leader     *User         // 负责人对象
}

type Menu struct {
    ID         string
    Name       string
    ParentID   string        // 父菜单 ID
    Parent     *Menu         // 父菜单对象
    Children   []*Menu       // 子菜单列表
    ApiIds     []string      // 关联的 API ID 列表
    Apis       []*API        // 关联的 API 对象列表
}
```

### 5. 仓库接口定义

```Go
// internal/domain/base/repo/user_repository.go
package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

type UserRepository interface {
    GetByID(id string) (*entity.User, error)           // ✅ ID 使用 string
    GetByUsername(userName string) (*entity.User, error)
    GetByWechatID(wxID string) (*entity.User, error)
    Save(user *entity.User) error
    Delete(id string) error                            // ✅ ID 使用 string
}
```

### 6. 持久化层实现

```Go
// internal/persistence/base/user_repository.go
package persistence

import (
    "fmt"
    "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"
    "github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
    "github.com/ix-pay/ixpay-pro/internal/infrastructure/persistence/database"
    "github.com/ix-pay/ixpay-pro/internal/persistence/common"
)

// Database Model（带 GORM 标签，使用雪花算法 ID）
type userModel struct {
    database.SnowflakeBaseModel
    Username      string `gorm:"size:50;not null;unique"`
    PasswordHash  string `gorm:"size:100;not null"`
    Email         string `gorm:"size:100"`
    Status        int    `gorm:"default:1"`
    PositionID    int64  `gorm:"index"`
    DepartmentID  int64  `gorm:"index"`
    WechatOpenID  string `gorm:"size:100;uniqueIndex;default:null"`
}

// TableName 指定表名
func (userModel) TableName() string {
    return "base_users"
}

// 转换为 Domain Entity（ID: int64 → string）
// ⭐ 详细转换逻辑已在 [核心架构 - Repository 层转换职责](#三 -1-repository-层是数据转换的核心) 中定义
func (m *userModel) toDomain() *entity.User {
    return &entity.User{
        ID:           common.ToString(m.ID),
        Username:     m.Username,
        Email:        m.Email,
        PositionID:   common.ToString(m.PositionID),
        DepartmentID: common.ToString(m.DepartmentID),
        CreatedBy:    common.ToString(m.CreatedBy),
    }
}

// 从 Domain Entity 转换（ID: string → int64）
// ⭐ 详细转换逻辑已在 [核心架构 - Repository 层转换职责](#三 -1-repository-层是数据转换的核心) 中定义
func fromDomain(user *entity.User) (*userModel, error) {
    id, createdBy, updatedBy := common.SetBaseFields(user.ID, user.CreatedBy, user.UpdatedBy)
    
    return &userModel{
        SnowflakeBaseModel: database.SnowflakeBaseModel{
            ID:        id,
            CreatedBy: createdBy,
            UpdatedBy: updatedBy,
        },
        Username:     user.Username,
        Email:        user.Email,
        PositionID:   common.TryParseInt64(user.PositionID),
        DepartmentID: common.TryParseInt64(user.DepartmentID),
    }, nil
}

// ⭐ GORM 标准做法说明
// 1. DeletedAt 字段：使用 gorm.DeletedAt 类型（GORM 标准）
//    - gorm.DeletedAt 是 sql.NullTime 的别名
//    - 包含 Valid (bool) 和 Time (time.Time) 两个字段
//    - GORM 自动处理软删除逻辑
// 2. 时间字段：使用 time.Time（不是 gorm.DateTime）
//    - CreatedAt/UpdatedAt：GORM 自动填充
//    - DeletedAt：GORM 自动管理软删除
// 3. 软删除操作：
//    - 删除：db.Delete(&model, id) - 自动设置 DeletedAt
//    - 查询：db.Find(&models) - 自动过滤 DeletedAt.Valid = true 的记录
//    - 恢复：db.Unscoped().Model(&model).Update("deleted_at", nil)
// 4. 检查软删除状态：
//    - if model.DeletedAt.Valid { /* 已删除 */ }
// 5. ⭐⭐⭐ Swagger 文档最佳实践 ⭐⭐⭐
//    - 数据库 Model 使用 gorm.DeletedAt（GORM 标准）
//    - API Response DTO 不要直接使用 Model，而是自定义 DTO 结构
//    - DTO 中只包含需要返回给前端的字段，不包含 DeletedAt 等内部字段
//    - 这样可以避免 Swagger 无法识别 gorm.DeletedAt 的问题
```

### 7. DTO 定义

```Go
// internal/dto/base/request/user.go
package request

type CreateUserRequest struct {
    Username string   `json:"userName" binding:"required"`
    Email    string   `json:"email" binding:"required,email"`
    Password string   `json:"password" binding:"required,min=6"`
    RoleIds  []string `json:"roleIds"`  // ✅ 使用 camelCase
}

// internal/dto/base/response/user.go
package response

type UserResponse struct {
    ID        string   `json:"id"`        // ✅ ID 使用 string
    Username  string   `json:"userName"`
    Email     string   `json:"email"`
    RoleIds   []string `json:"roleIds"`   // ✅ 使用 camelCase
    CreatedAt string   `json:"createdAt"` // ✅ 使用 camelCase
}
```

### 8. API Handler

```Go
// internal/app/base/api/user_handler.go
package api

import (
    "github.com/gin-gonic/gin"
    "github.com/ix-pay/ixpay-pro/internal/dto/base/request"
    "github.com/ix-pay/ixpay-pro/internal/dto/base/response"
    "github.com/ix-pay/ixpay-pro/internal/app/base/application"
)

type UserHandler struct {
    userAppService *application.UserAppService
}

func NewUserHandler(userAppService *application.UserAppService) *UserHandler {
    return &UserHandler{userAppService: userAppService}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
    var req request.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(err.Error(), c)
        return
    }
    
    if err := h.userAppService.CreateUser(&req); err != nil {
        response.Error(err.Error(), c)
        return
    }
    
    response.OkWithMessage("创建成功", c)
}
```

### 9. 依赖注入配置

```Go
// internal/app/base/wire.go
package base

import (
    "github.com/google/wire"
    
    // Domain 层
    "github.com/ix-pay/ixpay-pro/internal/domain/base/service"
    
    // Persistence 层
    "github.com/ix-pay/ixpay-pro/internal/persistence/base"
    
    // Application 层
    "github.com/ix-pay/ixpay-pro/internal/app/base/application"
    
    // API 层
    "github.com/ix-pay/ixpay-pro/internal/app/base/api"
)

// ProviderSetBase Base 模块的依赖注入集合
var ProviderSetBase = wire.NewSet(
    // Domain Service
    service.NewUserDomainService,
    service.NewRoleDomainService,
    
    // Repository 实现
    persistence.NewUserRepository,
    persistence.NewRoleRepository,
    
    // Application Service
    application.NewUserAppService,
    application.NewRoleAppService,
    
    // API Handler
    api.NewUserHandler,
    api.NewRoleHandler,
    
    // Base 应用实例
    NewAppBase,
)
```

### 10. 日志记录规范

**所有日志消息必须使用中文**

```Go
// ✅ 正确
logger.Info("用户登录成功", "user_id", userID, "ip", clientIP)
logger.Error("创建用户失败", "error", err)

// ❌ 错误
logger.Info("User login successful", "user_id", userID)
logger.Error("Failed to create user", "error", err)
```

### 11. GORM 关联表内联查询（Preload 懒加载）⭐

**使用 GORM 的** **`Preload`** **方法实现关联查询功能**

#### 11.1 数据库模型关联标签定义

**⭐⭐⭐ 关联字段命名规范（必须遵守）⭐⭐⭐**

**核心原则**：GORM 关联标签中的字段名必须与数据库表结构中的实际字段名完全一致（使用 snake\_case）

**多对一关联（Belongs To）**

```Go
type userModel struct {
    // ... 其他字段
    DepartmentID int64  `gorm:"index"`      // Go 字段名：PascalCase
    PositionID   int64  `gorm:"index"`      // Go 字段名：PascalCase
    
    // GORM 关联标签 - 多对一
    // ✅ 正确：foreignKey 和 references 使用数据库字段名（snake_case）
    Department *departmentModel `gorm:"foreignKey:department_id;references:id"`
    Position   *positionModel   `gorm:"foreignKey:position_id;references:id"`
    
    // ❌ 错误：不要使用 Go 字段名（PascalCase）
    // Department *departmentModel `gorm:"foreignKey:DepartmentID;references:ID"`
}
```

**多对多关联（Many To Many）**

```Go
type userModel struct {
    // ... 其他字段
    
    // GORM 关联标签 - 多对多（通过中间表 base_role_users）
    Roles []*roleModel `gorm:"many2many:base_role_users;joinForeignKey:user_id;joinReferences:role_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
```

**一对多关联（Has Many）**

```Go
type roleModel struct {
    // ... 其他字段
    
    // GORM 关联关系 - 一对多（通过中间表）
    Users []*userModel `gorm:"many2many:base_role_users;joinForeignKey:role_id;joinReferences:user_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
```

**关键字段说明**

| 参数               | 用途                   | 命名规则                  | 示例                           |
| ---------------- | -------------------- | --------------------- | ---------------------------- |
| `foreignKey`     | 指定**当前模型**中的外键字段     | snake\_case（数据库字段名）   | `department_id`, `parent_id` |
| `references`     | 指定**关联模型**的主键字段      | snake\_case（通常是 `id`） | `id`                         |
| `joinForeignKey` | 指定**当前模型**在中间表中的外键字段 | snake\_case（中间表字段名）   | `user_id`, `role_id`         |
| `joinReferences` | 指定**关联模型**在中间表中的外键字段 | snake\_case（中间表字段名）   | `role_id`, `menu_id`         |

#### 11.2 关联关系枚举定义（类型安全）⭐⭐⭐

**在接口层定义关联关系类型和常量**

```Go
// internal/domain/base/repo/user_repository.go
package repo

// UserRelation 用户关联关系类型（类型安全的枚举）
type UserRelation string

const (
    UserRelationDepartment UserRelation = "Department"  // 部门
    UserRelationPosition   UserRelation = "Position"    // 岗位
    UserRelationRoles      UserRelation = "Roles"       // 角色
)

// UserRepository 用户数据仓库接口
type UserRepository interface {
    // GetByID 根据 ID 查询用户并支持加载关联数据
    // relations 参数使用 UserRelation 类型，提供编译期类型检查
    GetByID(id string, relations ...UserRelation) (*entity.User, error)
    // ... 其他方法
}
```

**优势**：

- ✅ **编译期类型检查**：只能传入预定义的常量，避免拼写错误
- ✅ **IDE 自动提示**：输入 `repo.UserRelation` 后自动提示可用常量
- ✅ **易于维护**：集中定义，易于扩展
- ✅ **无循环依赖**：定义在接口层，符合依赖倒置原则

**使用示例**：

```Go
// internal/domain/base/service/user_service.go
func (s *UserService) GetUserInfo(userID string) (*entity.User, error) {
    // ✅ 正确：使用类型安全的常量
    user, err := s.repo.GetByID(userID, repo.UserRelationDepartment)
    
    // ❌ 错误：编译失败 - 不能直接使用字符串
    // user, err := s.repo.GetByID(userID, "Department")
    
    if err != nil {
        s.log.Error("获取用户信息失败", "error", err)
        return nil, err
    }
    return user, nil
}
```

#### 11.3 Repository 实现层

实现层将 UserRelation 类型转换为 string 用于 GORM：

```Go
// GetByID 根据 ID 查询用户并支持加载关联数据
func (r *userRepository) GetByID(id string, relations ...repo.UserRelation) (*entity.User, error) {
    intID, err := common.ParseInt64(id)
    if err != nil {
        return nil, err
    }

    var dbModel userModel
    query := r.db.Where("id = ?", intID)
    
    // 根据指定的关联关系进行 Preload
    for _, relation := range relations {
        query = query.Preload(string(relation))
    }
    
    result := query.First(&dbModel)
    if result.Error != nil {
        return nil, result.Error
    }

    return dbModel.toDomain(), nil
}
```

#### 11.4 Preload 查询用法

**基本用法**

```Go
// 只查基本信息（不加载关联）
var user userModel
db.First(&user, id)

// 加载单个关联
db.Preload("Department").First(&user, id)

// 加载多个关联
db.Preload("Department").Preload("Position").Preload("Roles").First(&user, id)

// 条件加载关联
db.Preload("Roles", "status = ?", 1).First(&user, id)

// 嵌套加载（加载关联的关联）
db.Preload("Department.Leader").Preload("Roles.Menus").First(&user, id)
```

**Service 层使用示例**

```Go
// 只查用户基本信息
user, err := repo.GetByID("123")

// 查用户 + 部门
user, err := repo.GetByID("123", repo.UserRelationDepartment)

// 查用户 + 部门 + 岗位 + 角色
user, err := repo.GetByID("123", repo.UserRelationDepartment, repo.UserRelationPosition, repo.UserRelationRoles)

// 分页查询带关联
users, total, err := repo.List(1, 10, nil, repo.UserRelationDepartment, repo.UserRelationRoles)
```

#### 11.5 toDomain 方法处理关联数据

核心要点：

- 在 `toDomain()` 方法中同时填充 ID 列表和完整对象
- 使用一个循环完成两个任务，避免多次遍历
- 按需加载：通过 Preload 控制是否加载完整对象

#### 11.6 Preload 的优势

1. **懒加载**：只在需要时加载关联数据，避免不必要的查询
2. **避免 N+1 问题**：批量加载关联数据，提高查询效率
3. **灵活控制**：可以精确控制加载哪些关联
4. **性能优化**：不会意外加载不需要的数据
5. **类型安全**：使用枚举常量，避免字符串拼写错误 ⭐

#### 11.7 注意事项

1. **关联字段类型**：
   - 多对一：使用指针类型 `*Model`
   - 多对多：使用切片类型 `[]Model`
2. **外键一致性**：外键字段类型必须一致（如都是 int64）
3. **空值处理**：关联可能为空，需要检查 nil
4. **中间表**：多对多关系需要中间表存在（如 base\_role\_users）
5. **性能考虑**：
   ```Go
   // ❌ 避免：在循环中查询（N+1 问题）
   for _, user := range users {
       db.First(&department, user.DepartmentID)
   }

   // ✅ 推荐：使用 Preload 批量加载
   db.Preload("Department").Find(&users)
   ```
6. **类型转换**：Repository 实现层需要将枚举类型转换为 string

#### 11.8 C# LINQ vs Go GORM 对比

**C# LINQ:**

```csharp
var user = context.Users
    .Include(u => u.Department)
    .Include(u => u.Position)
    .Include(u => u.Roles)
    .FirstOrDefault(u => u.Id == id);
```

**Go GORM:**

```Go
var user userModel
db.Preload("Department").
   Preload("Position").
   Preload("Roles").
   First(&user, id)
```

***

## 前端规范（Vue3 + TypeScript）

### 0. Element Plus + Tailwind CSS 混合架构规范

**核心原则**：

1. **混合架构**：Element Plus（交互组件）+ Tailwind CSS（布局样式）
2. **职责分离**：Element Plus 负责交互，Tailwind 负责视觉
3. **主题支持**：必须支持白天/黑暗模式切换
4. **样式隔离**：Element Plus 组件标签不使用 Tailwind CSS

**使用规范**：

**Tailwind CSS（布局与装饰）**

- ✅ 普通 HTML 标签使用 Tailwind CSS（flex、grid、spacing、颜色等）
- ✅ 暗黑模式使用 `dark:` 前缀

**Element Plus（交互组件）**

- ✅ 表格、表单、弹窗等使用 Element Plus
- ❌ **禁止在 Element Plus 组件标签上添加 Tailwind CSS 类**
- ✅ 使用 Element Plus 原生属性（width、height、size、type 等）

**禁止事项**：

- ❌ 禁止在 Element Plus 组件标签上使用 Tailwind CSS 类
- ❌ 禁止使用内联样式
- ❌ 禁止手写媒体查询
- ❌ 禁止自定义无意义类名

**示例代码**：

```Vue
<!-- ✅ 正确：普通 HTML 使用 Tailwind，Element Plus 使用原生属性 -->
<template>
  <div class="flex items-center gap-4 p-4">
    <el-button type="primary" size="large" @click="handleClick">
      点击
    </el-button>
    <el-table :data="tableData" style="width: 100%">
      <el-table-column prop="name" label="姓名" />
    </el-table>
  </div>
</template>

<!-- ❌ 错误：在 Element Plus 组件上使用 Tailwind CSS 类 -->
<template>
  <el-button type="primary" class="mt-4 p-2">  <!-- ❌ -->
    点击
  </el-button>
  <el-table :data="tableData" class="w-full">  <!-- ❌ -->
    <el-table-column prop="name" label="姓名" class="p-4" />  <!-- ❌ -->
  </el-table>
</template>
```

### 1. 代码格式

- 缩进：2 个空格
- 引号：单引号
- 分号：不使用分号
- 空行：逻辑块之间保留空行

### 2. 命名规范

- 组件名：PascalCase（如 UserProfile.vue）
- 变量/方法名：camelCase
- 常量：UPPER\_CASE
- TypeScript 接口：全字段 camelCase（如 userId、createdAt）
- **所有字段使用 camelCase**

```TypeScript
// ✅ 正确：camelCase
export interface UserInfo {
  id: string              // string 类型
  userName: string
  nickname: string
  email: string
  phone: string
  avatar: string
  status: number
  roleIds: string[]       // string 数组
  currentRoleId: string   // string 类型
  createdAt: string
  updatedAt: string
}

// ❌ 错误：snake_case
export interface UserInfo {
  id: string
  user_name: string       // ❌
  current_role_id: string // ❌
  created_at: string      // ❌
}
```

### 3. 组件结构

- 模板顺序：template → script setup → style scoped
- script 内部：imports → props → emits → 响应式数据 → 计算属性 → 方法 → 生命周期

### 4. TypeScript 规范

- props、emits、函数必须有类型注解
- 复杂类型使用接口定义

### 5. API 调用规范

```TypeScript
// ✅ 正确的 API 调用
const getUserInfo = async () => {
  const res = await request({
    url: '/api/admin/base/user/info',
    method: 'get',
  })
  
  // 使用 camelCase 字段
  const userInfo = res.data
  console.log(userInfo.id)              // string
  console.log(userInfo.currentRoleId)   // string
  console.log(userInfo.roleIds)         // string[]
  
  return userInfo
}
```

### 6. 表单提交规范

```TypeScript
// ✅ 前端传 string 类型 ID
const submitForm = async () => {
  const formData = {
    id: userId,           // string
    roleIds: roleIds,     // string[]
    status: '1',          // string（如果后端用 string 接收）
  }
  
  await updateUser(formData)
}

// ✅ 编辑时处理
const handleEdit = (user: UserInfo) => {
  return {
    id: String(user.id),
    roleIds: user.roleIds?.map(id => String(id)),
    status: String(user.status),
  }
}
```

### 7. 样式规范

- 必须使用 scoped 或 CSS Modules

***

## 典型场景流程

**标准 7 步开发流程已在** **[标准 7 步开发流程](#标准-7-步开发流程)** **中详细定义，以下是常见场景的简化流程和工作流程：**

### 场景 1：从零开始开发新功能

**示例**: "实现用户管理模块"

**简化流程**：按照 \[标准 7 步开发流程]\(#标准 -7-步开发流程) 执行，重点关注：

- 第 1 步：需求分析与 API 设计（用户 CRUD + 权限控制）
- 第 2 步：后端开发（生成脚手架 → 实现逻辑 → 测试）
- 第 4 步：前端开发（基于 API 接口）
- 第 5 步：前后端联调

### 场景 2：修改现有功能

**示例**: "用户列表增加搜索功能"

**简化流程**：

- 第 1 步：需求分析（明确修改：增加搜索参数）
- 第 2 步：API 设计（修改接口：增加 query 参数）
- 第 3 步：后端修改 → 测试 → 格式化
- 第 4 步：前端修改 → 联调测试
- 第 6 步：代码审查 → 修复
- 第 7 步：提交

### 场景 3：修复 Bug

**示例**: "用户删除报错"

**简化流程**：

- 问题定位 → 查看错误日志 → 分析原因
- 后端修复 → 生成测试 → 运行测试 → 格式化
- 后端验证 → 编译检查 → 代码审查
- 前端验证 → 手动测试
- 提交

### 场景 4：代码优化重构

**示例**: "优化用户服务代码结构"

**简化流程**：

- 分析 → 识别需要优化的代码
- 重构 → 生成测试 → 运行测试 → 格式化
- 审查 → 代码审查 → 设计原则检查
- 提交

### 工作流程

```
1. 识别用户意图（新增/修改功能）
   ↓
2. 自动检查前后端规范符合性
   ├─ 后端：JSON 标签命名、ID 类型、响应格式
   ├─ 前端：TypeScript 接口命名、ID 类型
   └─ 一致性：前后端字段名是否匹配
   ↓
3. 提供流程指导
   - 说明当前步骤
   - 提示使用技能
   - 检查关键点
   ↓
4. 执行相应操作
   ├─ 生成脚手架代码
   ├─ 格式化代码（go fmt / npm run format）
   ├─ 生成单元测试
   ├─ 执行构建/运行命令
   └─ 生成 API 文档
   ↓
5. 生成检查报告
   - ✅ 符合规范的项目
   - ⚠️ 需要整改的问题
   - 💡 优化建议
```

***

## 自动检查清单

### 后端检查

- [ ] JSON 标签是否使用 camelCase
- [ ] 所有 ID 字段在应用层和 DTO 中是否使用 string 类型
- [ ] 响应是否使用 `baseRes.OkWithDetailed()`
- [ ] 日志消息是否使用中文
- [ ] API 处理器是否没有直接操作数据库
- [ ] API 路径是否使用 `/api/[module]/base/`
- [ ] 是否有完整的单元测试
- [ ] 测试覆盖率是否 > 80%
- [ ] 依赖注入配置是否正确使用 wire
- [ ] 错误处理是否使用 fmt.Errorf 包装
- [ ] 仓库实现是否正确处理 ID 类型转换

### 前端检查

- [ ] TypeScript 接口字段是否使用 camelCase
- [ ] ID 字段类型是否为 `string` 或 `string[]`
- [ ] API 调用是否正确处理响应
- [ ] 表单提交是否使用正确的字段名
- [ ] API 路径是否使用 `/base/` 前缀
- [ ] 响应格式是否为 `{code, data, msg}`
- [ ] 列表字段是否使用 `items`
- [ ] 分页参数是否使用 `page, page_size`

### 前后端对齐检查

- [ ] 路径一致
- [ ] 请求参数一致
- [ ] 响应格式一致
- [ ] 错误处理一致
- [ ] 分页格式一致
- [ ] 数据类型一致（ID 使用 string）

***

## 使用示例

### 示例 1：新功能开发

**用户**: "实现用户管理模块"

**技能响应**:

```
好的，我将按照全栈开发流程来实现用户管理模块。

**当前步骤**: 第 1 步 - 需求分析与 API 设计

**功能分析**:
- 用户 CRUD 操作（创建、查询、更新、删除）
- 权限控制
- 分页列表

**API 设计**:
1. GET    /api/admin/base/user  - 用户列表
2. POST   /api/admin/base/user       - 创建用户
3. PUT    /api/admin/base/user/:id   - 更新用户
4. DELETE /api/admin/base/user/:id   - 删除用户
5. GET    /api/admin/base/user/:id   - 用户详情

现在开始第 2 步：后端开发
正在生成后端脚手架代码...
```

### 示例 2：修改功能

**用户**: "用户列表增加搜索功能"

**技能响应**:

```
好的，我将按照修改功能流程来实现搜索功能。

**当前步骤**: 第 1 步 - 需求分析

**修改内容**:
- 增加搜索参数：keyword, status
- 修改响应：过滤后的列表

**API 修改**:
GET /api/admin/base/user/list
新增参数：
- keyword (string, optional): 搜索关键词
- status (int, optional): 状态过滤

现在开始修改后端代码...
```

### 示例 3：检测并修复规范问题

**用户**: "我完成了用户列表 API 的开发"

**技能响应**:

```
✅ API 路径格式正确（/api/admin/base/users）
✅ 响应格式统一（使用 baseRes）
✅ 分层架构清晰（Handler -> Application Service -> Domain Service）
⚠️ 发现以下问题：

【问题 1】JSON 标签命名不规范
位置：internal/dto/base/response/user.go:20
当前：CreatedAt string `json:"created_at"`
建议：CreatedAt string `json:"createdAt"`
规范：JSON 标签应使用 camelCase

【问题 2】ID 字段未使用 string
位置：internal/domain/base/entity/user.go:10
当前：ID int64
建议：ID string
规范：领域实体 ID 应使用 string 类型

需要我帮你自动修改这些问题吗？
```

***

## 与其他技能的协同

本技能是**全栈开发的核心技能**，与其他技能协同工作：

```
全栈开发（流程指导 + 规范定义）
    ↓ 需要创建/修改文件时
项目目录结构指引（文件位置确定）
    ↓ 需要执行开发命令时
后端开发工具集（后端命令执行）
前端开发工具集（前端命令执行）
    ↓ 代码完成后
代码质量检查器（质量审查）
    ↓ 开发完成后
Git 提交与推送工具（提交代码）
```

### 技能调用规范

1. **文件操作**：当需要创建或修改文件时，必须先调用**项目目录结构指引**技能来确定正确的文件位置。
2. **命令执行**：当需要执行开发命令时，应调用对应的开发工具集：
   - 后端命令（如 `go build`、`swag init` 等）→ 调用**后端开发工具集**
   - 前端命令（如 `npm install`、`npm run dev` 等）→ 调用**前端开发工具集**
3. **规范检查**：代码完成后，应调用**代码质量检查器**进行规范审查。
4. **代码提交**：开发完成后，应调用**Git 提交与推送工具**工具提交代码。

***

## 菜单与按钮权限开发规范 ⭐⭐⭐

### 一、核心概念

#### 1. 菜单类型（type 字段）

- **type: 1** - 目录（不可打开页面的菜单，用于分组）
- **type: 2** - 菜单（可以打开页面的菜单，实际可访问的页面）
- **type: 3** - 按钮（按钮权限数据，不是菜单，是页面内的操作按钮）

#### 2. 权限标识命名规范

```
格式：{模块}:{资源}:{操作}

示例：
- system:user:add      - 系统管理 - 用户资源 - 新增操作
- system:user:edit     - 系统管理 - 用户资源 - 编辑操作
- system:user:delete   - 系统管理 - 用户资源 - 删除操作
- task:task:execute    - 任务管理 - 任务资源 - 执行操作
```

***

### 二、数据处理流程

#### 1. 核心流程

```
后端返回菜单数据 (ApiMenuItem[])
    ↓
步骤 1: 提取按钮权限 (type: 3) → 存储到 routerStore.buttonPermissions
    ↓
步骤 2: 过滤 type: 3 数据 → 在 filterRoutesByPermission() 和 formatRouter() 中过滤
    ↓
步骤 3: 权限检查过滤 → 检查用户是否有菜单权限
    ↓
生成菜单导航 (只包含 type: 1 和 type: 2)
```

#### 2. 关键代码位置

```TypeScript
// web/src/stores/modules/router.ts
// 按钮权限提取逻辑
const extractButtonPermissions = (menus: ApiMenuItem[]): string[] => {
  const buttons: string[] = []
  const traverse = (items: ApiMenuItem[]) => {
    for (const item of items) {
      if (item.type === 3 && item.permission) {
        buttons.push(item.permission)
      }
      if (item.children?.length > 0) {
        traverse(item.children)
      }
    }
  }
  traverse(menus)
  return buttons
}

// 在 filterRoutesByPermission() 和 formatRouter() 中过滤 type: 3 数据
if ((route as ApiMenuItem).type === 3) {
  return false  // 或 continue
}
```

***

### 三、前端按钮权限控制

#### 1. 使用方式

```Vue
<!-- 页面按钮使用 v-auth-btn 指令 -->
<el-button v-auth-btn="'system:user:add'" @click="handleAdd">
  新增
</el-button>

<el-button v-auth-btn="'system:user:edit'" @click="handleEdit">
  编辑
</el-button>

<el-button v-auth-btn="'system:user:delete'" @click="handleDelete">
  删除
</el-button>
```

#### 2. 权限控制逻辑

- ✅ **有权限**：按钮显示
- ❌ **无权限**：按钮从 DOM 中移除（不是隐藏）
- ⭐ **管理员**：自动拥有所有权限（`role === 'admin'`）

#### 3. 核心文件

- `web/src/directive/authBtn.ts` - 按钮权限指令
- `web/src/utils/permission.ts` - 权限检查函数 `hasButtonPermission()`
- `web/src/stores/modules/router.ts` - 按钮权限存储 `buttonPermissions`

***

### 四、后端菜单配置示例

```json
{
  "id": "xxx",
  "parentId": "父菜单 ID",
  "path": "user",
  "name": "UserManagement",
  "title": "用户管理",
  "type": 2,
  "children": [
    {
      "id": "xxx",
      "parentId": "xxx",
      "name": "UserAdd",
      "title": "新增用户",
      "type": 3,
      "permission": "system:user:add"
    },
    {
      "id": "xxx",
      "parentId": "xxx",
      "name": "UserEdit",
      "title": "编辑用户",
      "type": 3,
      "permission": "system:user:edit"
    }
  ]
}
```

***

### 五、开发流程

#### 新增页面按钮权限的 3 个步骤

**步骤 1: 后端配置菜单数据**

在 `internal/domain/base/seed/menu_seed.go` 中添加按钮权限数据：

```Go
// 在用户管理菜单下创建按钮权限
_, err = ms.createOrGetMenu(logger, &entity.Menu{
    ParentID:   userMenu.ID,  // 父菜单 ID
    Name:       "UserAdd",
    Title:      "新增用户",
    Icon:       "Plus",
    Sort:       1,
    Status:     1,
    Type:       entity.MenuTypeButton,  // type: 3 - 按钮权限
    Permission: "system:user:add",      // 权限标识
})
if err != nil {
    return err
}
```

**步骤 2: 前端页面使用指令**

```Vue
<el-button v-auth-btn="'system:user:add'">新增</el-button>
```

**步骤 3: 测试验证**

- 有权限账号 → 按钮显示
- 无权限账号 → 按钮隐藏
- 管理员账号 → 所有按钮显示

***

### 六、注意事项

1. **type: 3 数据不应该出现在菜单导航中**
   - 只用于提取权限标识
   - 在路由过滤时移除
2. **按钮权限数据应该被提取并存储**
   - 存储到 `routerStore.buttonPermissions`
   - 用于权限检查
3. **前后端权限验证**
   - ✅ 前端：按钮级别权限控制（提升体验）
   - ✅ 后端：API 接口权限验证（安全保障）
   - ❌ 禁止：只依赖前端，后端不验证
4. **管理员特权**
   - 管理员角色 (`role === 'admin'`) 自动拥有所有权限
   - 不需要检查具体的权限标识

***

## 最佳实践

1. **小步快跑**
   - ✅ 推荐：创建一个 API → 测试 → 提交
   - ❌ 不推荐：一次性创建所有 API，最后统一提交
2. **测试先行**
   - ✅ 推荐：先写测试用例 → 实现功能 → 重构优化
   - ❌ 不推荐：先实现功能，有时间再写测试
3. **文档同步**
   - ✅ 推荐：API 修改后立即更新文档
   - ❌ 不推荐：代码改了文档不改
4. **代码审查**
   - ✅ 推荐：每次提交前自动审查
   - ❌ 不推荐：不审查直接提交
5. **分层架构**
   - ✅ 推荐：严格遵循 DDD 分层架构
   - ❌ 不推荐：跨层调用或混合职责
6. **依赖注入**
   - ✅ 推荐：使用 wire 进行依赖注入
   - ❌ 不推荐：手动创建依赖实例
7. **按钮权限控制** ⭐
   - ✅ 推荐：使用 `v-auth-btn` 指令控制按钮显示
   - ✅ 推荐：后端配置 type: 3 的按钮权限数据
   - ✅ 推荐：前端自动提取并存储按钮权限
   - ❌ 不推荐：将 type: 3 数据显示在菜单导航中
   - ❌ 不推荐：只依赖前端控制，后端不验证权限

***

## 常见问题解答

**Q1: 应该先开发后端还是前端？**
A: 推荐后端优先：先设计 API → 实现后端 → 测试接口 → 开发前端 → 联调

**Q2: 测试覆盖率多少合适？**
A: 至少 80%：核心业务 > 80%，一般业务 > 60%，简单 CRUD > 40%

**Q3: 什么时候使用代码审查？**
A: 每次提交前：新增功能、修改功能、修复 Bug、重构代码后都要审查

**Q4: 如何保证前后端数据类型一致？**
A: 统一规范：ID 使用 string，时间使用 ISO8601 字符串，数字使用 number，布尔值使用 bool

**Q5: 如何处理 ID 类型转换？**
A: Repository 层负责 `string ↔ int64` 的转换，应用层和 DTO 层使用 string 类型

***

