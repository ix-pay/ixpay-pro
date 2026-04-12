<p align="center">
  <img src="./docs/images/ixpay.png" width="300" height="300" alt="IXPay Pro Logo" />
</p>

English | [简体中文](./README.md)

# IXPay Pro

A high-performance payment API service based on Go language and Gin framework, focusing on providing WeChat payment solutions, integrating user authentication, payment processing, and task management functions.

## Project Structure

```
ixpay-pro/
├── common/             # Common modules (documentation, standards, etc.)
├── gxy/                # Gateway service (Go)
├── h5app/              # H5 application frontend (in development)
├── server/             # Backend service (Go)
├── weapp/              # WeChat mini-program frontend (in development)
└── web/                # Vue3 + TypeScript frontend
```

## Tech Stack

### Backend (server)

- **Language**: Go 1.24.6
- **Web Framework**: Gin v1.10.1
- **Dependency Injection**: Wire v0.7.0
- **Database**: PostgreSQL + GORM v1.30.3
- **Cache**: Redis v9.13.0
- **Authentication**: JWT v5.3.0
- **Configuration**: Viper v1.20.1
- **Logging**: Zap v1.27.0

### Frontend (web)

- **Framework**: Vue 3
- **Language**: TypeScript
- **UI Library**: Element Plus v2.11.2
- **State Management**: Pinia v3.0.3
- **Router**: Vue Router v4.5.1
- **Build Tool**: Vite v7.0.6

## Core Features

### User Authentication

- Registration, login, WeChat login
- Token refresh and permission management

### Payment Processing

- Create payment, query payment
- Cancel payment and handle WeChat payment notifications

### Task Management

- Add, remove, start, stop, and retry tasks

## Quick Start

### Backend Deployment

#### Docker Deployment (Recommended)

```bash
cd server
# Create .env file and configure environment variables
cp .env.example .env
# Start service
docker-compose up -d
```

#### Local Running

```bash
# Enter backend service directory
cd server

# Install dependencies
go mod download

# Generate dependency injection code
wire ./internal/app

# Generate API documentation (execute in server directory)
swag init -g cmd/ixpay-pro/main.go --output ./docs --parseInternal --parseDependency

# Run application
GO_ENV=development go run cmd/ixpay-pro/main.go

# Access API documentation
http://127.0.0.1:8586/swagger/index.html

# Build executable
go build -o ./build/ixpay-pro.exe cmd/ixpay-pro/main.go
```

### Frontend Running

```bash
cd web
# Install dependencies
npm install
# Development mode
npm run serve
# Production build
npm run build
```

## API Documentation

After the application starts, you can access Swagger API documentation at:

```
http://127.0.0.1:8586/swagger/index.html
```

## Contributing

1. Fork the repository
2. Create your feature branch: `git checkout -b feature/AmazingFeature`
3. Commit your changes: `git commit -m 'Add some AmazingFeature'`
4. Push to the branch: `git push origin feature/AmazingFeature`
5. Open a Pull Request

## License

IXPay Pro is released under the Apache License 2.0.