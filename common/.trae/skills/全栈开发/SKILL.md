---
name: 全栈开发
description: 进行全栈开发时触发，包括新功能开发、功能修改、Bug 修复、代码重构等场景，提供从需求分析到提交部署的完整 7 步流程指导，确保前后端规范统一
---

## 核心原则：前后端交互规范 ⭐

### 1. JSON 序列化命名规范

**统一使用 camelCase（小驼峰命名）**

- ✅ **正确**: `userId`, `createdAt`, `currentRoleId`
- ❌ **错误**: `user_id` (snake_case), `UserID` (PascalCase)

**原因**:
- 前端 TypeScript/JavaScript 使用 camelCase
- 保持前后端命名一致性
- 符合行业标准

### 2. ID 字段处理规范

**所有 ID 字段在应用层和 DTO 中使用 string 类型**

- ✅ **正确**: `"id": "696243340630298624"`
- ❌ **错误**: `"id": 696243340630298624`

**原因**: JavaScript Number 精度限制（±2^53），Go int64 会丢失精度

***

## 核心架构：数据流转与转换规范 ⭐⭐⭐

### 一、完整的数据流转路径

#### 1. 写入流程（API → Database）

```
HTTP 请求 → API Handler → Domain Service → Repository → Database
          (DTO 层)      (领域实体)    (数据模型)   (数据库)
```

**详细步骤**:

```
┌─────────────────────────────────────────────────────────┐
│ 1. HTTP POST /api/admin/base/user                      │
│    Body: {"username": "admin", "email": "admin@ix.com"}│
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│ 2. API Handler (internal/app/base/api/user_handler.go) │
│    - ShouldBindJSON → RegisterRequest DTO              │
│    - 验证参数：binding:"required,email"                │
│    - 调用 service.Register(req.Username, req.Email)    │
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│ 3. Domain Service (internal/domain/base/service/)      │
│    - 业务逻辑：密码加密、邮箱验证                       │
│    - 创建 Domain Entity                                 │
│    - 调用 repo.Create(user)                            │
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│ 4. Repository (internal/persistence/base/)             │
│    - ⭐ 转换：Domain Entity → Database Model            │
│    - fromDomain(): string → int64 (ID 转换)            │
│    - 调用 db.Create(dbModel)                           │
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│ 5. PostgreSQL 数据库                                    │
│    - INSERT INTO base_users (id, username, email)      │
│    - 返回：雪花算法 ID (int64)                         │
└─────────────────────────────────────────────────────────┘
```

#### 2. 查询流程（Database → API）

```
PostgreSQL → Database Model → Domain Entity → API Handler → Response DTO → HTTP 响应
   (int64)      toDomain()     (string)      手动构建      (string)
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
│    - 返回：*entity.User（Domain Entity）               │
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│ 4. Repository (internal/persistence/base/)             │
│    - string → int64: ParseInt(id)                      │
│    - SQL: SELECT * FROM base_users WHERE id = ?        │
│    - 返回：userModel (Database Model)                  │
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│ 5. PostgreSQL 数据库                                     │
│    - 返回：数据库记录（ID: int64）                      │
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│ 6. Repository toDomain()                               │
│    - ⭐ int64 → string: FormatInt(m.ID, 10)            │
│    - 返回：*entity.User（ID: string）                  │
└─────────────────────────────────────────────────────────┘
                        ↓ (逐层返回)
┌─────────────────────────────────────────────────────────┐
│ 7. API Handler 构建 DTO                                 │
│    - ⭐ 手动映射：entity.User → UserInfoResponse       │
│    - 角色转换：Role Entity → RoleInfo DTO              │
└─────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────┐
│ 8. HTTP Response                                        │
│    {                                                    │
│      "code": 200,                                       │
│      "data": {                                          │
│        "id": "1234567890",  ✅ string                  │
│        "username": "admin",                             │
│        "roles": [{"id": "1", "name": "管理员"}]         │
│      }                                                  │
│    }                                                    │
└─────────────────────────────────────────────────────────┘
```

### 二、转换层次总结表

| 转换阶段 | 位置 | 转换方向 | 转换方法 | 示例代码位置 |
|---------|------|---------|---------|-------------|
| **HTTP → DTO** | API Handler | JSON → Request DTO | `ctx.ShouldBindJSON()` | [`user_handler.go`](d:\g\ixpay-pro\server\internal\app\base\api\user_handler.go) |
| **DTO → 参数** | API Handler | Request DTO → 简单参数 | 直接传递字段 | [`user_handler.go`](d:\g\ixpay-pro\server\internal\app\base\api\user_handler.go) |
| **参数 → Domain Entity** | Domain Service | 简单参数 → Entity | 构造器创建 | [`user_service.go`](d:\g\ixpay-pro\server\internal\domain\base\service\user_service.go) |
| **Domain Entity → Database Model** ⭐ | Repository | Entity → Model | [`fromDomain()`](d:\g\ixpay-pro\server\internal\persistence\base\user_repository.go#L64-L90) | [`user_repository.go:64`](d:\g\ixpay-pro\server\internal\persistence\base\user_repository.go#L64) |
| **Database Model → Domain Entity** ⭐ | Repository | Model → Entity | [`toDomain()`](d:\g\ixpay-pro\server\internal\persistence\base\user_repository.go#L37-L61) | [`user_repository.go:37`](d:\g\ixpay-pro\server\internal\persistence\base\user_repository.go#L37) |
| **Domain Entity → DTO** | API Handler | Entity → Response DTO | 手动构建 | [`user_handler.go`](d:\g\ixpay-pro\server\internal\app\base\api\user_handler.go) |

### 三、关键设计原则

#### 1. Repository 层是数据转换的核心 ⭐⭐⭐

**所有** `Database Model ↔ Domain Entity` 的转换都在 Repository 层完成：

```go
// internal/persistence/base/user_repository.go

// ⭐ Database Model → Domain Entity（查询时使用）
func (m *userModel) toDomain() *entity.User {
    return &entity.User{
        ID:            common.ToString(m.ID),      // ✅ int64 → string
        Username:      m.Username,
        PositionID:    common.ToString(m.PositionID),  // ✅ int64 → string
        DepartmentID:  common.ToString(m.DepartmentID), // ✅ int64 → string
        CreatedBy:     common.ToString(m.CreatedBy),    // ✅ int64 → string
    }
}

// ⭐ Domain Entity → Database Model（写入时使用）
func fromDomain(user *entity.User) (*userModel, error) {
    id, createdBy, updatedBy := common.SetBaseFields(user.ID, user.CreatedBy, user.UpdatedBy)
    
    return &userModel{
        SnowflakeBaseModel: database.SnowflakeBaseModel{
            ID:        id,
            CreatedBy: createdBy,
            UpdatedBy: updatedBy,
        },
        Username:     user.Username,
        PositionID:   common.TryParseInt64(user.PositionID),
        DepartmentID: common.TryParseInt64(user.DepartmentID),
    }, nil
}
```

**Repository 方法中的转换示例**:

```go
// 查询：GetByID
func (r *userRepository) GetByID(id string) (*entity.User, error) {
    // 1️⃣ string → int64（API ID → 数据库 ID）
    intID, err := strconv.ParseInt(id, 10, 64)
    
    // 2️⃣ 查询数据库
    var dbModel userModel
    result := r.db.Where("id = ?", intID).First(&dbModel)
    
    // 3️⃣ Database Model → Domain Entity
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
    user.ID = common.ToString(dbModel.ID)
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

```go
// internal/domain/base/service/user_service.go

// ✅ 正确：接收简单参数，返回 Domain Entity
func (s *UserService) Register(username, password, email string) (*entity.User, error) {
    // 业务逻辑：密码加密
    passwordHash := encryption.GeneratePasswordHash(password)
    
    // 创建 Domain Entity
    user := &entity.User{
        Username:     username,
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

```go
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

| 层级 | ID 类型 | 原因 |
|------|--------|------|
| **数据库存储** | `int64` | ✅ 索引性能好，存储空间小，使用雪花算法 |
| **Domain Entity** | `string` | ✅ 避免 JSON 精度丢失，前后端统一 |
| **DTO** | `string` | ✅ API 传输安全，JavaScript 友好 |

**转换责任**: Repository 层负责 `string ↔ int64` 的转换

```go
// 数据库层：int64
type userModel struct {
    database.SnowflakeBaseModel  // ✅ 使用雪花算法 ID
}

// 领域层：string
type User struct {
    ID string  // ✅ 领域层使用 string
}

// DTO 层：string
type UserResponse struct {
    ID string `json:"id"`  // ✅ DTO 使用 string
}
```

#### 5. 错误处理规范

```go
// ✅ 正确：使用 fmt.Errorf 包装错误，保留错误链
func (s *UserService) GetUserByID(id string) (*entity.User, error) {
    user, err := s.userRepo.GetByID(id)
    if err != nil {
        return nil, fmt.Errorf("获取用户失败 (ID=%s): %w", id, err)
    }
    return user, nil
}

// ❌ 错误：直接返回原始错误，丢失上下文
func (s *UserService) GetUserByID(id string) (*entity.User, error) {
    user, err := s.userRepo.GetByID(id)
    if err != nil {
        return nil, err  // ❌ 丢失上下文
    }
    return user, nil
}
```

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

### 0. Go 代码风格规范

#### 代码格式

- 4 空格缩进
- 每行≤120 字符
- K&R 括号风格
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

```go
// ✅ 正确：camelCase
type UserResponse struct {
    ID            string   `json:"id"`
    Username      string   `json:"username"`
    CreatedAt     string   `json:"createdAt"`
    CurrentRoleId string   `json:"currentRoleId"`
}

// ❌ 错误：snake_case
type UserResponse struct {
    ID            string   `json:"id"`
    Username      string   `json:"username"`
    CreatedAt     string   `json:"created_at"`        // ❌
    CurrentRoleId string   `json:"current_role_id"`   // ❌
}
```

### 4. 领域实体定义

```go
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
```go
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

| 设计方式 | 性能 | 灵活性 | 内存占用 | 适用场景 |
|---------|------|--------|---------|---------|
| **只保留 ID 列表** | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ | 只需判断关联关系 |
| **只保留完整对象** | ⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ | 需要完整信息展示 |
| **两者都保留** ⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | **推荐：适应多场景** |

**使用场景**：

1. **登录场景**（使用 ID 列表）：
```go
// 登录时只需要角色 ID 生成 JWT token
user, err := s.repo.GetByID(userID)  // 不加载 Roles
if len(user.RoleIds) > 0 {
    roleCode = user.RoleIds[0]  // ✅ 使用 RoleIds，轻量高效
}
```

2. **用户详情场景**（使用完整对象）：
```go
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

3. **业务判断场景**（使用 ID 列表）：
```go
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
```go
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
```go
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

```go
// internal/domain/base/repo/user_repository.go
package repo

import "github.com/ix-pay/ixpay-pro/internal/domain/base/entity"

type UserRepository interface {
    GetByID(id string) (*entity.User, error)           // ✅ ID 使用 string
    GetByUsername(username string) (*entity.User, error)
    GetByWechatID(wxID string) (*entity.User, error)
    Save(user *entity.User) error
    Delete(id string) error                            // ✅ ID 使用 string
}
```

### 6. 持久化层实现

```go
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

```go
// internal/dto/base/request/user.go
package request

type CreateUserRequest struct {
    Username string   `json:"username" binding:"required"`
    Email    string   `json:"email" binding:"required,email"`
    Password string   `json:"password" binding:"required,min=6"`
    RoleIds  []string `json:"roleIds"`  // ✅ 使用 camelCase
}

// internal/dto/base/response/user.go
package response

type UserResponse struct {
    ID        string   `json:"id"`        // ✅ ID 使用 string
    Username  string   `json:"username"`
    Email     string   `json:"email"`
    RoleIds   []string `json:"roleIds"`   // ✅ 使用 camelCase
    CreatedAt string   `json:"createdAt"` // ✅ 使用 camelCase
}
```

### 8. API Handler

```go
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

```go
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

### 10. 错误处理规范

```go
// ✅ 正确：使用 fmt.Errorf 包装错误，保留错误链
func (s *UserService) GetUserByID(id string) (*entity.User, error) {
    user, err := s.userRepo.GetByID(id)
    if err != nil {
        return nil, fmt.Errorf("获取用户失败 (ID=%s): %w", id, err)
    }
    return user, nil
}

// ❌ 错误：直接返回原始错误，丢失上下文
func (s *UserService) GetUserByID(id string) (*entity.User, error) {
    user, err := s.userRepo.GetByID(id)
    if err != nil {
        return nil, err  // ❌ 丢失上下文
    }
    return user, nil
}
```

### 11. 日志记录规范

**所有日志消息必须使用中文**

```go
// ✅ 正确
logger.Info("用户登录成功", "user_id", userID, "ip", clientIP)
logger.Error("创建用户失败", "error", err)

// ❌ 错误
logger.Info("User login successful", "user_id", userID)
logger.Error("Failed to create user", "error", err)
```

### 13. GORM 关联表内联查询（Preload 懒加载）⭐

**使用 GORM 的 `Preload` 方法实现类似 C# LINQ 的关联查询功能**

#### 13.1 数据库模型关联标签定义

**多对一关联（Belongs To）**
```go
type userModel struct {
    // ... 其他字段
    DepartmentID int64  `gorm:"index"`
    PositionID   int64  `gorm:"index"`
    
    // GORM 关联标签 - 多对一
    Department *departmentModel `gorm:"foreignKey:DepartmentID;references:ID"`
    Position   *positionModel   `gorm:"foreignKey:PositionID;references:ID"`
}
```

**多对多关联（Many To Many）**
```go
type userModel struct {
    // ... 其他字段
    
    // GORM 关联标签 - 多对多（通过中间表 base_role_users）
    Roles []roleModel `gorm:"many2many:base_role_users;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
```

**一对多关联（Has Many）**
```go
type roleModel struct {
    // ... 其他字段
    
    // GORM 关联关系 - 一对多（通过中间表）
    Users []userModel `gorm:"many2many:base_role_users;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
```

#### 13.2 ⭐ 关联关系枚举定义（类型安全）⭐⭐⭐

**在接口层定义关联关系类型和常量**

```go
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
```go
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

#### 13.3 Repository 实现层

**实现层将 UserRelation 类型转换为 string 用于 GORM**

```go
// internal/persistence/base/user_repository.go
package persistence

import (
    "github.com/ix-pay/ixpay-pro/internal/domain/base/repo"
    // ... 其他导入
)

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
        query = query.Preload(string(relation))  // ✅ 转换为字符串用于 GORM
    }
    
    result := query.First(&dbModel)
    if result.Error != nil {
        return nil, result.Error
    }

    return dbModel.toDomain(), nil
}
```

#### 13.4 Preload 查询用法

**基本用法**
```go
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
```go
// 只查用户基本信息
user, err := repo.GetByID("123")

// 查用户 + 部门
user, err := repo.GetByID("123", repo.UserRelationDepartment)

// 查用户 + 部门 + 岗位 + 角色
user, err := repo.GetByID("123", repo.UserRelationDepartment, repo.UserRelationPosition, repo.UserRelationRoles)

// 分页查询带关联
users, total, err := repo.List(1, 10, nil, repo.UserRelationDepartment, repo.UserRelationRoles)
```

#### 13.5 toDomain 方法处理关联数据

```go
func (m *userModel) toDomain() *entity.User {
    user := &entity.User{
        ID:       common.ToString(m.ID),
        Username: m.Username,
        // ... 其他基本字段
    }
    
    // 处理关联数据 - 部门
    if m.Department != nil {
        user.Department = m.Department.toDomain()
    }
    
    // 处理关联数据 - 岗位
    if m.Position != nil {
        user.Position = m.Position.toDomain()
    }
    
    // 处理关联数据 - 角色
    if len(m.Roles) > 0 {
        roleIDs := make([]string, len(m.Roles))
        for i, role := range m.Roles {
            roleIDs[i] = common.ToString(role.ID)
        }
        user.RoleIds = roleIDs
    }
    
    return user
}
```

#### 13.6 Preload 的优势

1. **懒加载**：只在需要时加载关联数据，避免不必要的查询
2. **避免 N+1 问题**：批量加载关联数据，提高查询效率
3. **灵活控制**：可以精确控制加载哪些关联
4. **性能优化**：不会意外加载不需要的数据
5. **类型安全**：使用枚举常量，避免字符串拼写错误 ⭐

#### 13.7 注意事项

1. **关联字段类型**：
   - 多对一：使用指针类型 `*Model`
   - 多对多：使用切片类型 `[]Model`
   
2. **外键一致性**：外键字段类型必须一致（如都是 int64）

3. **空值处理**：关联可能为空，需要检查 nil
   ```go
   if user.Department != nil {
       // 使用部门数据
   }
   ```

4. **中间表**：多对多关系需要中间表存在（如 base_role_users）

5. **性能考虑**：
   ```go
   // ❌ 避免：在循环中查询
   for _, user := range users {
       db.First(&department, user.DepartmentID)  // N+1 问题
   }
   
   // ✅ 推荐：使用 Preload 批量加载
   db.Preload("Department").Find(&users)
   ```

6. **类型转换**：Repository 实现层需要将枚举类型转换为 string
   ```go
   // ✅ 正确：在实现层转换
   query.Preload(string(relation))
   ```

#### 13.8 C# LINQ vs Go GORM 对比

**C# LINQ:**
```csharp
var user = context.Users
    .Include(u => u.Department)
    .Include(u => u.Position)
    .Include(u => u.Roles)
    .FirstOrDefault(u => u.Id == id);
```

**Go GORM:**
```go
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

```vue
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
- 常量：UPPER_CASE
- TypeScript 接口：全字段 camelCase（如 userId、createdAt）
- **所有字段使用 camelCase**

```typescript
// ✅ 正确：camelCase
export interface UserInfo {
  id: string              // string 类型
  username: string
  nickname: string
  email: string
  phone: string
  avatar: string
  status: number
  roleIds: string[]       // string 数组
  currentRoleId: string   // camelCase
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

```typescript
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

```typescript
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

### 场景 1：从零开始开发新功能

**示例**: "实现用户管理模块"

**流程**:

1. 需求分析 → 明确功能：用户 CRUD + 权限控制
2. API 设计 → 设计 4 个接口（GET/POST/PUT/DELETE）
3. 后端开发 → 生成脚手架 → 实现逻辑 → 编写测试 → 格式化 → 生成文档
4. 后端验证 → 编译检查 → 运行测试 → API 测试 → 代码审查
5. 前端开发 → 生成组件 → 实现页面 → 联调测试
6. 前后端联调 → 启动服务 → 联调测试 → 修复问题
7. 代码审查 → 全面审查 → 自动修复 → 手动修复
8. 提交部署 → 格式化 → 更新文档 → 提交代码

### 场景 2：修改现有功能

**示例**: "用户列表增加搜索功能"

**流程**:

1. 需求分析 → 明确修改：增加搜索参数
2. API 设计 → 修改接口：增加 query 参数
3. 后端修改 → 手动修改 → 生成测试 → 运行测试 → 格式化
4. 后端验证 → 编译检查 → 代码审查
5. 前端修改 → 手动修改 → 联调测试
6. 代码审查 → 规范检查 → 自动修复
7. 提交 → 格式化 → 提交代码

### 场景 3：修复 Bug

**示例**: "用户删除报错"

**流程**:

1. 问题定位 → 查看错误日志 → 分析原因
2. 后端修复 → 手动修复 → 生成测试 → 运行测试 → 格式化
3. 后端验证 → 编译检查 → 代码审查
4. 前端验证 → 手动测试
5. 提交 → 格式化 → 提交代码

### 场景 4：代码优化重构

**示例**: "优化用户服务代码结构"

**流程**:

1. 分析 → 识别需要优化的代码
2. 重构 → 手动重构 → 生成测试 → 运行测试 → 格式化
3. 审查 → 代码审查 → 设计原则检查
4. 提交 → 格式化 → 提交代码

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

## 工作流程

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
1. GET    /api/admin/base/user/list  - 用户列表
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
