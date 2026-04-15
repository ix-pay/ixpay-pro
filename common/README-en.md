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
| **Language** | Go | 1.24.6 | Core development language, providing high performance and concurrency capabilities |
| **Web Framework** | Gin | v1.10.1 | Lightweight HTTP service framework, providing routing, middleware, etc. |
| **Dependency Injection** | Wire | v0.7.0 | Compile-time dependency injection tool, improving code maintainability |
| **Database** | PostgreSQL | 13+ | Powerful open-source relational database, supporting complex queries and transactions |
| | GORM | v1.30.3 | Feature-rich ORM library, simplifying database operations |
| **Cache** | Redis | 6+ | High-performance key-value storage, used for caching and session management |
| **Authentication** | JWT | v5.3.0 | Stateless identity authentication token, supporting cross-service authentication |
| **Configuration** | Viper | v1.20.1 | Flexible configuration file management tool, supporting multiple configuration formats |
| **Logging** | Zap | v1.27.0 | High-performance structured logging library, supporting multiple log levels |
| **Task Scheduling** | Cron | v3.0.1 | Scheduled task scheduling library, used for executing periodic tasks |
| **API Documentation** | Swagger | - | Automatic API documentation generation tool, facilitating interface debugging |
| **Monitoring** | Prometheus | - | Open-source monitoring system, used for system performance monitoring |
| **Rate Limiting** | golang.org/x/time/rate | - | API rate limiting library, preventing system overload |
| **Snowflake Algorithm** | Snowflake | - | Distributed ID generation algorithm, ensuring data uniqueness |
| **Captcha** | captcha | - | Captcha generation and verification library, improving system security |

### Frontend (web)

| Category | Technology/Framework | Version | Description |
|----------|---------------------|---------|-------------|
| **Framework** | Vue 3 | - | Modern frontend framework, providing reactive data binding and component-based development |
| **UI Library** | Element Plus | v2.11.2 | Vue 3-based UI component library, providing rich interface elements |
| **Language** | TypeScript | - | Static type checking, improving code quality and maintainability |
| **Build Tool** | Vite | v7.0.6 | Modern frontend build tool, providing fast development experience |
| **State Management** | Pinia | v3.0.3 | Vue 3 official recommended state management library |
| **Router** | Vue Router | v4.5.1 | Vue official router library, implementing single-page application navigation |
| **HTTP Client** | Axios | - | Promise-based HTTP client, used for API calls |

### Gateway (gxy)

- **Language**: Go (pure standard library)
- **Core Functions**: Service registration and discovery, load balancing (round-robin algorithm), health checks, cluster data synchronization, request proxy forwarding
- **Technical Features**: High performance, lightweight, thread-safe

## Core Features

### Backend Management Functions

- **🔐 User Authentication**: Supports registration, login, WeChat login, and token refresh, using JWT for identity verification
- **👮 Permission Management**: RBAC+ABAC hybrid model-based permission management, supporting menu, API route, and button-level permission control
- **👥 Role Management**: Role CRUD, permission assignment, role inheritance, and permission group management
- **📋 Menu Management**: Menu CRUD, tree structure management, supporting dynamic menu generation
- **⚙️ Configuration Management**: System configuration CRUD, supporting multi-environment configurations
- **📚 Dictionary Management**: Dictionary table and dictionary item management, supporting data classification and standardization
- **📝 Operation Logs**: Records user operation logs, supporting log query and analysis
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
- **🔒 Security Protection**: Built-in input validation, SQL injection prevention, XSS attack prevention, and other security measures
- **⚡ Performance Optimization**: Uses Redis caching, database indexing, and other technologies to optimize system performance
- **📦 Containerized Deployment**: Supports Docker containerized deployment, simplifying deployment and operations

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
| Go | 1.20+ | Backend development language, recommended version 1.24.6 |
| Node.js | 16+ | Frontend development environment, recommended version 18.x |
| npm | 8+ | Frontend dependency management, recommended version 9.x |
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
- **ixpay-server**: Backend service, port 8586
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
go build -o ixpay-server cmd/ixpay-pro/main.go
./ixpay-server

# Access API documentation
http://127.0.0.1:8586/swagger/index.html
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

- **Swagger UI**: http://localhost:8586/swagger/index.html
- **API Documentation JSON**: http://localhost:8586/swagger/doc.json
- **API Documentation YAML**: http://localhost:8586/swagger/doc.yaml

### API Interface Classification

#### Authentication and Authorization APIs

| Interface Name | Method | Path | Description |
|---------------|--------|------|-------------|
| Register | POST | /api/admin/auth/register | User registration |
| Login | POST | /api/admin/auth/login | User login |
| Captcha | POST | /api/admin/auth/captcha | Get captcha |
| Refresh Token | POST | /api/admin/auth/refresh-token | Refresh access token |
| Logout | POST | /api/admin/auth/logout | User logout |

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

#### Role Management APIs

| Interface Name | Method | Path | Description |
|---------------|--------|------|-------------|
| Create Role | POST | /api/admin/roles | Create new role |
| Get Role Details | GET | /api/admin/roles/detail | Get role details |
| Update Role | PUT | /api/admin/roles | Update role info |
| Delete Role | DELETE | /api/admin/roles | Delete role |
| Get Role List | GET | /api/admin/roles | Get role list |
| Assign Users to Role | POST | /api/admin/roles/assign-users | Assign users to role |
| Assign Menus to Role | POST | /api/admin/roles/assign-menus | Assign menus to role |

#### Payment Management APIs

| Interface Name | Method | Path | Description |
|---------------|--------|------|-------------|
| Create Payment | POST | /api/payment | Create payment order |
| Query Payment | GET | /api/payment/{id} | Query payment details |
| Get User Payment List | GET | /api/payment | Get user payment list |
| Cancel Payment | PUT | /api/payment/{id}/cancel | Cancel payment order |

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
| SERVER_PORT | Service port | 8586 |
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
  port: 8586            # Service port
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
- **ixpay-server**: Backend service, port 8586
- **postgres**: PostgreSQL database, port 5432
- **redis**: Redis cache, port 6379

### Building Docker Images

```bash
cd server
# Build image
docker build -t ixpay-server .

# Run container
docker run -d --name ixpay-server \
  -p 8586:8586 \
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
   curl http://localhost:8586/health
   
   # Test login interface
   curl -X POST http://localhost:8586/api/admin/auth/login \
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