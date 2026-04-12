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

// 写入：Create
func (r *userRepository) Create(user *entity.User) error {
    // 1️⃣ Domain Entity → Database Model
    dbModel, err := fromDomain(user)
    
    // 2️⃣ 保存到数据库
    return r.db.Create(dbModel).Error
}
```

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
git 提交推送（提交代码）
```

### 技能调用规范

1. **文件操作**：当需要创建或修改文件时，必须先调用**项目目录结构指引**技能来确定正确的文件位置。
2. **命令执行**：当需要执行开发命令时，应调用对应的开发工具集：
   - 后端命令（如 `go build`、`swag init` 等）→ 调用**后端开发工具集**
   - 前端命令（如 `npm install`、`npm run dev` 等）→ 调用**前端开发工具集**
3. **规范检查**：代码完成后，应调用**代码质量检查器**进行规范审查。
4. **代码提交**：开发完成后，应调用**git 提交推送**工具提交代码。

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
