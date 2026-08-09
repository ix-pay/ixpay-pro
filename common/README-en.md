<p align="center">
  <img src="./docs/images/ixpay.png" width="300" height="300" alt="IXPay Pro Logo" />
</p>

English | [简体中文](./README.md)

# IXPay Pro

IXPay Pro is a high-performance payment management system based on Go language and Gin framework, focusing on providing WeChat payment solutions. The system adopts a front-end and back-end separation architecture and DDD (Domain-Driven Design) layered architecture, integrating a complete backend management system, payment processing, and task management functions.

## Project Structure

```
ixpay-pro/
├── common/             # Common modules (documentation, standards, skills, etc.)
├── gxy/                # Gateway service (Go, pure standard library)
├── h5app/              # H5 application frontend (initialized)
├── miniapp/            # Mini-program frontend (initialized)
├── server/             # Backend service (Go, DDD architecture)
├── weapp/              # Enterprise WeChat H5 application (initialized)
└── web/                # Vue3 + TypeScript frontend management backend
```

## Tech Stack

### Backend (server)

| Category | Technology/Framework | Version | Description |
|----------|---------------------|---------|-------------|
| **Language** | Go | 1.26.5 | Core development language, providing high performance and concurrency capabilities |
| **Web Framework** | Gin | v1.11.0 | Lightweight HTTP service framework, providing routing, middleware, etc. |
| **Dependency Injection** | Wire | v0.7.0 | Compile-time dependency injection tool, improving code maintainability |
| **Database** | PostgreSQL | 13+ | Powerful open-source relational database, supporting complex queries and transactions |
| | GORM | v1.31.1 | Feature-rich ORM library, simplifying database operations |
| **Cache** | Redis (go-redis) | v9.17.2 | High-performance key-value storage, used for caching and session management |
| **Authentication** | JWT | v5.3.0 | Stateless identity authentication token, supporting cross-service authentication |
| **Configuration** | Viper | v1.21.0 | Flexible configuration file management tool, supporting multiple configuration formats |
| **Logging** | Zap | v1.27.1 | High-performance structured logging library, supporting multiple log levels |
| **Task Scheduling** | Cron | v3.0.1 | Scheduled task scheduling library, used for executing periodic tasks |
| **System Monitoring** | gopsutil | v4.26.2 | System resource monitoring library, obtaining CPU, memory, disk, etc. |
| **Monitoring** | Prometheus | v1.23.2 | Open-source monitoring system, used for system performance monitoring |
| **Rate Limiting** | golang.org/x/time/rate | v0.14.0 | API rate limiting library, preventing system overload |
| **Circuit Breaker** | gobreaker | v1.0.0 | Service circuit breaker library, improving system fault tolerance |
| **API Documentation** | Swagger | v1.16.6 | Automatic API documentation generation tool, facilitating interface debugging |
| **Snowflake Algorithm** | Snowflake | - | Distributed ID generation algorithm, ensuring data uniqueness |
| **Captcha** | base64Captcha | v1.3.8 | Captcha generation and verification library, improving system security |

### Frontend (web)

| Category | Technology/Framework | Version | Description |
|----------|---------------------|---------|-------------|
| **Framework** | Vue 3 | v3.5.18 | Modern frontend framework, providing reactive data binding and component-based development |
| **UI Library** | Element Plus | v2.11.2 | Vue 3-based UI component library, providing rich interface elements |
| **Language** | TypeScript | v5.8.0 | Static type checking, improving code quality and maintainability |
| **Build Tool** | Vite | v7.0.6 | Modern frontend build tool, providing fast development experience |
| **State Management** | Pinia | v3.0.3 | Vue 3 official recommended state management library |
| **Router** | Vue Router | v4.5.1 | Vue official router library, implementing single-page application navigation |
| **HTTP Client** | Axios | v1.12.2 | Promise-based HTTP client, used for API calls |
| **Chart Library** | ECharts | v6.0.0 | Data visualization chart library |
| **CSS Framework** | Tailwind CSS | v3.4.17 | Utility-first CSS framework |
| **Linting** | ESLint | v9.31.0 | JavaScript/TypeScript code linting tool |
| **Formatting** | Prettier | v3.6.2 | Code formatting tool |

### Gateway (gxy)

- **Language**: Go (pure standard library)
- **Core Functions**: Service registration and discovery, load balancing (round-robin algorithm), health checks, cluster data synchronization, request proxy forwarding
- **Technical Features**: High performance, lightweight, thread-safe

## Core Features

### Backend Management Functions

- **🔐 User Authentication**: Supports registration, login, WeChat login, and token refresh, using JWT for identity verification
- **👮 Permission Management**: RBAC+ABAC hybrid model-based permission management, supporting menu, API route, and button-level permission control
- **👥 Role Management**: Role CRUD, permission assignment, role inheritance, and permission management
- **📋 Menu Management**: Menu CRUD, tree structure management, supporting dynamic menu generation
- **⚙️ Configuration Management**: System configuration CRUD, supporting multi-environment configurations
- **📚 Dictionary Management**: Dictionary table and dictionary item management, supporting data classification and standardization
- **🏢 Department Management**: Organization department management with tree structure and hierarchical relationships
- **📌 Position Management**: Position information CRUD, supporting user-position association
- **📢 Notice Management**: System notice publication and management with reading record tracking
- **📝 Operation Logs**: Records user operation logs, supporting log query and analysis
- **🔑 Login Logs**: Records user login logs, supporting login behavior analysis
- **🌱 Seed Data Management**: System initialization data management, ensuring rapid deployment and configuration

### Payment Functions

- **💳 Payment Processing**: Supports creating payments, querying payments, canceling payments, and handling WeChat payment notifications
- **📱 WeChat Payment**: Integrates WeChat Payment API, supporting QR code payment, H5 payment, and other payment methods
- **💰 Transaction Management**: Payment transaction query, statistics, and analysis

### System Functions

- **📄 Documentation**: Integrates Swagger API documentation, facilitating interface debugging and integration
- **🛑 Graceful Shutdown**: Supports signal processing and graceful shutdown, ensuring stable service exit
- **🆔 Distributed ID**: Integrates Snowflake algorithm to generate unique IDs, ensuring data consistency
- **🔑 Captcha Service**: Supports generating and verifying captchas, improving system security
- **🌐 CORS Support**: Built-in CORS middleware, solving cross-domain issues in front-end and back-end separation architecture
- **📊 Monitoring System**: Supports Prometheus monitoring and Zap logging, ensuring stable system operation
- **📈 System Monitoring**: Real-time monitoring of system CPU, memory, disk, and other system resources
- **👥 Online Users**: Real-time online user monitoring and management
- **🔄 Circuit Breaker Protection**: Integrated gobreaker circuit breaker, improving system fault tolerance
- **🔒 Security Protection**: Built-in input validation, SQL injection prevention, XSS attack prevention, and other security measures
- **⚡ Performance Optimization**: Uses Redis caching, database indexing, and other technologies to optimize system performance
- **📦 Containerized Deployment**: Supports Docker containerized deployment, simplifying deployment and operations
- **⏰ Task Scheduling**: Powerful task scheduling system, supporting cache tasks, database tasks, HTTP tasks, script tasks, and other task types

### Gateway Functions

- **Service Registration and Discovery**: Automatically registers and manages backend service instances
- **Load Balancing**: Request distribution based on round-robin algorithm
- **Health Checks**: Real-time monitoring of backend service health status
- **Cluster Synchronization**: Supports multi-gateway node data synchronization
- **Request Proxy**: Efficient HTTP request forwarding

## Quick Start

### Environment Requirements

| Component | Version Requirement | Usage |
|-----------|-------------------|-------|
| Go | 1.26.5 | Backend development language |
| Node.js | 20.19+ or 22.12+ | Frontend development environment |
| npm | 9+ | Frontend dependency management |
| PostgreSQL | 13+ | Relational database, recommended version 14.x |
| Redis | 6+ | Cache, session management, recommended version 7.x |
| Docker | 20.10+ | Containerized deployment (optional) |
| Docker Compose | 1.29+ | Container orchestration (optional) |

### Backend Deployment

#### Docker Deployment (Recommended)

```bash
cd server
# Create .env file and configure environment variables
cp .env.example .env
# Start service
docker-compose up -d
```

This will start the following services:
- **ixpay-server**: Backend service, port 8081
- **postgres**: PostgreSQL database, port 5432
- **redis**: Redis cache, port 6379

#### Local Running

```bash
# Enter backend service directory
cd server

# Install dependencies
go mod download
go mod tidy

# Configure database and Redis
# Edit configs/config.yaml file

# Generate dependency injection code
wire ./internal/app

# Generate API documentation (execute in server directory)
swag init -g cmd/ixpay-pro/main.go --output ./docs --parseInternal --parseDependency

# Run database migrations
go run cmd/ixpay-pro/main.go migrate

# Run seed data
go run cmd/ixpay-pro/main.go seed

# Run application
# Development mode
go run cmd/ixpay-pro/main.go

# Production mode
go build -o ./cmd/build/ixpay-server cmd/ixpay-pro/main.go
./ixpay-server

# Access API documentation
http://127.0.0.1:8081/swagger/index.html
```

### Frontend Running

```bash
# Enter frontend directory
cd web

# Install dependencies
npm install

# Development mode
npm run serve

# Production build
npm run build
```

### Gateway Running

```bash
# Enter gateway directory
cd gxy

# Install dependencies
go mod download

# Run gateway
go run cmd/gateway/main.go

# Build executable
go build -o gateway cmd/gateway/main.go
```

## API Documentation

### Generate API Documentation

Use Swagger to generate API documentation:

```bash
swag init -g cmd/ixpay-pro/main.go --output ./docs --parseDependency --parseInternal --parseDepth 1
```

### Access API Documentation

The system integrates Swagger/OpenAPI documentation, accessible after starting the service:

- **Swagger UI**: http://localhost:8081/swagger/index.html
- **API Documentation JSON**: http://localhost:8081/swagger/doc.json
- **API Documentation YAML**: http://localhost:8081/swagger/doc.yaml

### API Interface Classification

#### Authentication and Authorization APIs

| Interface Name | Method | Path | Description |
|---------------|--------|------|-------------|
| Register | POST | /api/admin/auth/register | User registration |
| Login | POST | /api/admin/auth/login | User login |
| Captcha | POST | /api/admin/auth/captcha | Get captcha |
| Refresh Token | POST | /api/admin/auth/refresh-token | Refresh access token |
| Logout | POST | /api/admin/auth/logout | User logout |
| JWT Blacklist | POST | /api/admin/auth/jwt/jsonInBlacklist | Add JWT to blacklist |

#### User Management APIs

| Interface Name | Method | Path | Description |
|---------------|--------|------|-------------|
| Get User Info | GET | /api/admin/user/info | Get current user info |
| Update User Info | PUT | /api/admin/user/info | Update user info |
| Get User List | GET | /api/admin/user | Get user list |
| Add User | POST | /api/admin/user | Add new user |
| Delete User | DELETE | /api/admin/user/:id | Delete user |
| Change Password | PUT | /api/admin/user/password | Change user password |
| Reset Password | PUT | /api/admin/user/reset-password | Reset user password |
| Switch Role | POST | /api/admin/user/switch-role | Switch user role |
| Get User Settings | GET | /api/admin/user/get-user-settings | Get user personal settings |
| Update User Settings | PUT | /api/admin/user/update-user-settings | Update user personal settings |
| Set User Authority | POST | /api/admin/user/setUserAuthority | Set user authority (single role) |
| Set User Authorities | POST | /api/admin/user/setUserAuthorities | Set user authorities (multiple roles) |

#### Role Management APIs

| Interface Name | Method | Path | Description |
|---------------|--------|------|-------------|
| Create Role | POST | /api/admin/role | Create new role |
| Get Role List | GET | /api/admin/role | Get role list |
| Get Role Detail | GET | /api/admin/role/:id/detail | Get role detail |
| Get Available APIs | GET | /api/admin/role/:id/available-apis | Get available APIs for role |
| Get Available Menus | GET | /api/admin/role/:id/available-menus | Get available menus for role |
| Save Role Permissions | POST | /api/admin/role/:id/permissions | Save role permissions |
| Update Role | PUT | /api/admin/role | Update role info (ID from request body) |
| Delete Role | DELETE | /api/admin/role | Delete role (ID from request body) |
| Get All Roles | GET | /api/admin/role/all | Get all roles |
| Assign Users to Role | POST | /api/admin/role/assign-users | Assign users to role |
| Assign Menus to Role | POST | /api/admin/role/assign-menus | Assign menus to role |
| Assign APIs to Role | POST | /api/admin/role/assign-api-routes | Assign API routes to role |

#### Menu Management APIs

| Interface Name | Method | Path | Description |
|---------------|--------|------|-------------|
| Get User Menu Tree | GET | /api/admin/menu | Get user menu tree |
| Create Menu | POST | /api/admin/menu | Create new menu |
| Update Menu | PUT | /api/admin/menu | Update menu info |
| Delete Menu | DELETE | /api/admin/menu/:id | Delete menu |
| Get Menu Page | GET | /api/admin/menu/page | Get menu paginated list |
| Get Menu Tree | GET | /api/admin/menu/tree | Get full menu tree |
| Get Delete Impact | GET | /api/admin/menu/:id/delete-impact | Get menu delete impact assessment |

#### API Management APIs

| Interface Name | Method | Path | Description |
|---------------|--------|------|-------------|
| Get API List | GET | /api/admin/apis | Get API route list |
| Get API Detail | GET | /api/admin/apis/:id | Get API route detail |
| Create API | POST | /api/admin/apis | Create API route |
| Update API | PUT | /api/admin/apis/:id | Update API route |
| Delete API | DELETE | /api/admin/apis/:id | Delete API route |

#### Department Management APIs

| Interface Name | Method | Path | Description |
|---------------|--------|------|-------------|
| Create Department | POST | /api/admin/dept | Create new department |
| Get Department List | GET | /api/admin/dept | Get department list |
| Get Department Detail | GET | /api/admin/dept/:id | Get department detail |
| Update Department | PUT | /api/admin/dept | Update department info (ID from request body) |
| Delete Department | DELETE | /api/admin/dept/:id | Delete department |
| Get Department Tree | GET | /api/admin/dept/tree | Get department tree |
| Update Department Leader | PUT | /api/admin/dept/:id/leader | Update department leader |

#### Position Management APIs

| Interface Name | Method | Path | Description |
|---------------|--------|------|-------------|
| Create Position | POST | /api/admin/position | Create new position |
| Get Position List | GET | /api/admin/position | Get position list |
| Get Position Detail | GET | /api/admin/position/:id | Get position detail |
| Update Position | PUT | /api/admin/position | Update position info (ID from request body) |
| Delete Position | DELETE | /api/admin/position/:id | Delete position |
| Get All Positions | GET | /api/admin/position/all | Get all positions |

#### Dictionary Management APIs

| Interface Name | Method | Path | Description |
|---------------|--------|------|-------------|
| Get Dict List | GET | /api/admin/dict | Get dictionary list |
| Create Dict | POST | /api/admin/dict | Create dictionary |
| Get Dict Detail | GET | /api/admin/dict/:id | Get dictionary detail |
| Update Dict | PUT | /api/admin/dict | Update dictionary (ID from request body) |
| Delete Dict | DELETE | /api/admin/dict/:id | Delete dictionary |
| Get Dict Items | GET | /api/admin/dict/items | Get dict items by dict ID |
| Create Dict Item | POST | /api/admin/dict/item | Create dict item |
| Get Active Items | GET | /api/admin/dict/code/:code/active-items | Get active items by code |

#### Configuration Management APIs

| Interface Name | Method | Path | Description |
|---------------|--------|------|-------------|
| Get Config List | GET | /api/admin/config | Get config list |
| Create Config | POST | /api/admin/config | Create config |
| Get Config Detail | GET | /api/admin/config/:id | Get config detail |
| Update Config | PUT | /api/admin/config | Update config |
| Delete Config | DELETE | /api/admin/config/:id | Delete config |
| Get Config By Key | GET | /api/admin/config/key | Get config by key |
| Get Active Configs | GET | /api/admin/config/active | Get all active configs |

#### Notice Management APIs

| Interface Name | Method | Path | Description |
|---------------|--------|------|-------------|
| Get Notice List | GET | /api/admin/notices | Get notice list |
| Create Notice | POST | /api/admin/notices | Create notice |
| Get Notice Detail | GET | /api/admin/notices/:id | Get notice detail |
| Update Notice | PUT | /api/admin/notices | Update notice (ID from request body) |
| Delete Notice | DELETE | /api/admin/notices/:id | Delete notice |
| Publish Notice | POST | /api/admin/notices/:id/publish | Publish notice |
| Mark as Read | POST | /api/admin/notices/:id/read | Mark notice as read |
| Check Is Read | GET | /api/admin/notices/:id/is-read | Check if notice is read |
| Get Statistics | GET | /api/admin/notices/statistics | Get notice statistics |

#### Operation Log APIs

| Interface Name | Method | Path | Description |
|---------------|--------|------|-------------|
| Get Log List | GET | /api/admin/logs | Get operation log list |
| Get Log Detail | GET | /api/admin/logs/:id | Get operation log detail |
| Delete Log | DELETE | /api/admin/logs/:id | Delete operation log |
| Batch Delete Logs | POST | /api/admin/logs/batch-delete | Batch delete operation logs |
| Get Log Statistics | GET | /api/admin/logs/statistics | Get operation log statistics |
| Clear Logs | POST | /api/admin/logs/clear | Clear operation logs by time range |

#### Login Log APIs

| Interface Name | Method | Path | Description |
|---------------|--------|------|-------------|
| Get Login Log List | GET | /api/admin/login-log | Get login log list |
| Get Login Log Detail | GET | /api/admin/login-log/:id | Get login log detail |
| Get Statistics | GET | /api/admin/login-log/statistics | Get login statistics |
| Get Abnormal Logins | GET | /api/admin/login-log/abnormal | Get abnormal login records |
| Record Login | POST | /api/admin/login-log | Record login log |
| Batch Delete | POST | /api/admin/login-log/batch-delete | Batch delete login logs |
| Clear Login Logs | POST | /api/admin/login-log/clear | Clear login logs |

#### Online User APIs

| Interface Name | Method | Path | Description |
|---------------|--------|------|-------------|
| Get Online User List | GET | /api/admin/online-user | Get online user list |
| Get Online User Detail | GET | /api/admin/online-user/:user_id | Get online user detail |
| Get Online Count | GET | /api/admin/online-user/count | Get online user count |
| Check Is Online | GET | /api/admin/online-user/online | Check if user is online |
| Force Offline | DELETE | /api/admin/online-user/:user_id | Force user offline |
| Batch Force Offline | POST | /api/admin/online-user/batch | Batch force users offline |

#### System Monitor APIs

| Interface Name | Method | Path | Description |
|---------------|--------|------|-------------|
| Get System Monitor | GET | /api/admin/monitor/system | Get system resource monitor |
| Get Cache Monitor | GET | /api/admin/monitor/cache | Get cache monitor |
| Get Database Monitor | GET | /api/admin/monitor/database | Get database monitor |
| Get Redis Keys | GET | /api/admin/monitor/redis-keys | Query Redis keys |
| Get Slow Queries | GET | /api/admin/monitor/slow-queries | Query slow query logs |

#### Task Management APIs

| Interface Name | Method | Path | Description |
|---------------|--------|------|-------------|
| Get Task List | GET | /api/admin/task | Get task list |
| Create Task | POST | /api/admin/task | Create task |
| Get Task Detail | GET | /api/admin/task/:id | Get task detail |
| Update Task | PUT | /api/admin/task/:id | Update task |
| Delete Task | DELETE | /api/admin/task/:id | Delete task |
| Start Task | POST | /api/admin/task/:id/start | Start task |
| Stop Task | POST | /api/admin/task/:id/stop | Stop task |
| Retry Task | POST | /api/admin/task/:id/retry | Retry failed task |
| Enable Task | POST | /api/admin/task/:id/enable | Enable task |
| Disable Task | POST | /api/admin/task/:id/disable | Disable task |
| Get Execution Logs | GET | /api/admin/task/:id/execution-logs | Get task execution logs |
| Search Execution Logs | GET | /api/admin/task/execution-logs | Search execution logs |
| Get Statistics | GET | /api/admin/task/statistics | Get task statistics |
| Get Dashboard | GET | /api/admin/task/dashboard | Get task dashboard |
| Set Task Group | POST | /api/admin/task/:id/group | Set task group |

#### Permission Audit Log APIs

| Interface Name | Method | Path | Description |
|---------------|--------|------|-------------|
| Get Permission Log List | GET | /api/admin/permission-logs | Get permission log list |
| Get Role Permission Logs | GET | /api/admin/permission-logs/roles/:roleId | Get role permission logs |

#### Node Management APIs

| Interface Name | Method | Path | Description |
|---------------|--------|------|-------------|
| Get Node List | GET | /api/admin/nodes | Get node list |
| Get Node Detail | GET | /api/admin/nodes/:id | Get node detail |
| Offline Node | POST | /api/admin/nodes/:id/offline | Offline node |
| Get Node Statistics | GET | /api/admin/nodes/statistics | Get node statistics |

#### Gateway Service Management APIs

| Interface Name | Method | Path | Description |
|---------------|--------|------|-------------|
| Get Service List | GET | /api/admin/gateway/services | Get gateway service list |

#### Payment Management APIs

| Interface Name | Method | Path | Description |
|---------------|--------|------|-------------|
| Create Payment | POST | /api/payment | Create payment order |
| Query Payment | GET | /api/payment/:id | Query payment details |
| Get User Payment List | GET | /api/payment | Get user payment list |
| Cancel Payment | PUT | /api/payment/:id/cancel | Cancel payment order |

### API Documentation Notes

- **Request Format**: All API interfaces support JSON format request bodies
- **Response Format**: All API interfaces return JSON format responses
- **Authentication**: Uses JWT token for authentication, add `Authorization: Bearer <token>` in request headers
- **Error Handling**: Unified error response format, including error code and error message
- **Pagination**: List interfaces support `page` and `page_size` parameters for pagination

## System Architecture

### Architecture Layers

1. **Frontend Layer**: Modern user interface built with Vue3 + Element Plus
2. **API Layer**: RESTful API interfaces based on Gin framework
3. **Service Layer**: Service components implementing core business logic
4. **Data Access Layer**: Data repositories interacting with databases
5. **Infrastructure Layer**: Provides authentication, caching, logging, and other basic services

### Module Division

- **Base Management Module** (`server/internal/app/base`): User, role, permission, menu, and other core management functions
- **WeChat Payment Module** (`server/internal/app/wx`): WeChat payment-related function implementations
- **Infrastructure Module** (`server/internal/infrastructure`): Authentication, caching, logging, database, and other basic services
- **Gateway Module** (`gxy`): Service registration and discovery, load balancing, health checks

### Technical Features

- **Modular Design**: Clear layered architecture, facilitating expansion and maintenance
- **RESTful API**: Follows RESTful design specifications, providing standardized interfaces
- **Permission System**: RBAC+ABAC hybrid model-based permission management
- **Caching Mechanism**: Uses Redis to cache permission information and hot data
- **Middleware**: Implements authentication, permission verification, operation logging, and other middleware
- **Dependency Injection**: Uses Wire for compile-time dependency injection, improving code maintainability
- **Unified Error Handling**: Implements global error handling mechanism
- **Comprehensive Logging**: Uses Zap for high-performance logging

## Configuration Notes

### Environment Variables

IXPay Pro supports configuring the system through environment variables, which override corresponding settings in configuration files:

| Variable Name | Description | Default Value |
|--------------|-------------|---------------|
| LOG_LEVEL | Log level (debug/info/warn/error) | info |
| SERVER_PORT | Service port | 8081 |
| SERVER_MODE | Server running mode (debug/release/test) | debug |
| JWT_SECRET | JWT secret key | Randomly generated |
| JWT_EXPIRE | JWT expiration time (seconds) | 3600 |
| REDIS_HOST | Redis host | localhost |
| REDIS_PORT | Redis port | 6379 |
| REDIS_PASSWORD | Redis password | "" |
| REDIS_DB | Redis database number | 0 |
| DB_HOST | Database host | localhost |
| DB_PORT | Database port | 5432 |
| DB_USER | Database user | ixpay |
| DB_PASSWORD | Database password | ixpay123 |
| DB_NAME | Database name | ixpay_pro |
| DB_SSLMODE | Database SSL mode | disable |

### Configuration Files

Main configuration file located at `server/configs/config.yaml`, including the following main parts:

```yaml
# Server configuration
server:
  port: 8081            # Service port
  mode: "debug"         # Running mode: debug, release, test

# Database configuration
database:
  type: "postgres"      # Database type
  host: "localhost"     # Database host
  port: 5432            # Database port
  user: "ixpay"         # Database user
  password: "ixpay123"   # Database password
  dbname: "ixpay_pro"   # Database name
  sslmode: "disable"    # SSL mode

# Redis configuration
redis:
  host: "localhost"     # Redis host
  port: 6379            # Redis port
  password: ""          # Redis password
  db: 0                 # Redis database number

# JWT configuration
jwt:
  secret: "your-secret-key"  # JWT secret key
  expire: 3600            # Expiration time (seconds)

# Logging configuration
logging:
  level: "info"          # Log level
  file: "logs/"          # Log file directory
```

## Docker Deployment

### Using Docker Compose

IXPay Pro provides complete Docker Compose configuration to start all services with one command:

```bash
cd server
docker-compose up -d
```

This will start the following services:
- **ixpay-server**: Backend service, port 8081
- **postgres**: PostgreSQL database, port 5432
- **redis**: Redis cache, port 6379

### Building Docker Images

```bash
cd server
# Build image
docker build -t ixpay-server .

# Run container
docker run -d --name ixpay-server \
  -p 8081:8081 \
  -e DB_HOST=postgres \
  -e DB_PORT=5432 \
  -e DB_USER=ixpay \
  -e DB_PASSWORD=ixpay123 \
  -e DB_NAME=ixpay_pro \
  -e REDIS_HOST=redis \
  -e REDIS_PORT=6379 \
  ixpay-server
```

## Troubleshooting

### Common Issues and Solutions

#### Database Connection Failure

- **Symptom**: Database connection error when starting service
- **Solutions**:
  - Check if PostgreSQL service is running
  - Verify database connection parameters (host, port, userName, password)
  - Check if database user has correct permissions
  - Verify if database is created

#### Redis Connection Failure

- **Symptom**: Redis connection error when starting service
- **Solutions**:
  - Check if Redis service is running
  - Verify Redis connection parameters (host, port, password)
  - Check if Redis has a password set
  - Verify network connectivity

#### JWT Authentication Failure

- **Symptom**: API request returns 401 Unauthorized error
- **Solutions**:
  - Check if JWT secret key is consistent
  - Verify if token is expired
  - Check if token format is correct
  - Verify Authorization header in request

#### Cross-Origin Issues

- **Symptom**: Cross-origin error when frontend requests backend API
- **Solutions**:
  - Check if API address configured in frontend is correct
  - Verify if CORS middleware is configured in backend
  - Check browser console error messages

#### Service Startup Failure

- **Symptom**: Service fails to start or exits immediately after starting
- **Solutions**:
  - Check if port is occupied
  - Verify configuration files are correct
  - Check error messages in log files

### Log Viewing

Backend logs are saved in `server/logs/` directory by default, classified by level:

- `error.log`: Error logs, recording system errors and exceptions
- `warn.log`: Warning logs, recording system warnings
- `info.log`: Information logs, recording system running status
- `debug.log`: Debug logs, recording detailed debug information

### Debugging Methods

1. **Enable Debug Mode**:
   - Change `server.mode` in configuration file to `debug`
   - Set environment variable `SERVER_MODE=debug`

2. **View Detailed Logs**:
   - Change `logging.level` in configuration file to `debug`
   - Set environment variable `LOG_LEVEL=debug`

3. **Test API with curl**:
   ```bash
   # Test health check interface
   curl http://localhost:8081/health
   
   # Test login interface
   curl -X POST http://localhost:8081/api/admin/auth/login \
     -H "Content-Type: application/json" \
     -d '{"userName": "admin", "password": "password123"}'
   ```

4. **Check Database Status**:
   ```bash
   # Connect to PostgreSQL
   psql -h localhost -U ixpay -d ixpay_pro
   
   # View table structure
   \dt
   
   # View data
   SELECT * FROM users;
   ```

## Security Recommendations

### Production Environment Configuration

1. **Basic Security Configuration**
   - Change default passwords, use strong password policy
   - Enable HTTPS, configure SSL certificates
   - Configure firewall rules, restrict access ports
   - Restrict API access IPs, use whitelist

2. **Server Security**
   - Regularly update operating system and software
   - Disable unnecessary services and ports
   - Configure secure SSH access
   - Use key authentication, disable password login

### Database Security

1. **Database Configuration**
   - Use strong passwords, change regularly
   - Restrict database access IPs
   - Principle of least privilege, assign appropriate permissions to database users
   - Enable database audit logging

2. **Data Protection**
   - Regularly backup database, develop recovery plan
   - Encrypt sensitive data at rest
   - Regularly clean expired data
   - Implement data access control

### Application Security

1. **Code Security**
   - Regularly update dependencies, fix security vulnerabilities
   - Enable input validation, prevent malicious input
   - Prevent SQL injection, use parameterized queries
   - Prevent XSS attacks, filter user input

2. **Authentication and Authorization**
   - Use secure password hashing algorithms
   - Implement multi-factor authentication
   - Regularly rotate JWT keys
   - Monitor abnormal login behavior

3. **API Security**
   - Implement API rate limiting, prevent brute force attacks
   - Use HTTPS to protect API communication
   - Verify identity and permissions for all API requests
   - Log API access logs

## Contributing

We welcome contributions in all forms!

1. **Fork the repository**
2. **Create your feature branch**: `git checkout -b feature/AmazingFeature`
3. **Commit your changes**: `git commit -m 'feat: add some AmazingFeature'`
4. **Push to the branch**: `git push origin feature/AmazingFeature`
5. **Open a Pull Request**

### Code Style

- **Backend**: Follow [Go Code Style and Development Guidelines](.trae/rules/Go 代码风格与开发规范.md)
- **Frontend**: Follow [Vue Code Style Guidelines](.trae/rules/Vue 代码风格规范.md)
- **Commit Messages**: Follow Conventional Commits specification

### Development Workflow

1. **Clone the repository**: `git clone https://github.com/ix-pay/ixpay-pro.git`
2. **Install dependencies**: Install dependencies according to each sub-project's README
3. **Configure environment**: Configure database, Redis, and other environment
4. **Develop features**: Develop new features in corresponding modules
5. **Write tests**: Write unit tests for new features
6. **Submit code**: Ensure all tests pass before submitting code

## License

IXPay Pro is released under the Apache License 2.0.

## Contact

- **Project Homepage**: https://github.com/ix-pay/ixpay-pro
- **Issue Tracker**: https://github.com/ix-pay/ixpay-pro/issues
- **Email**: support@ixpay.pro

---

<p align="center">Made with ❤️ by IXPay Pro Team</p>