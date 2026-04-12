# 第一阶段：构建阶段
FROM golang:1.24.6-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git gcc musl-dev wire

# 复制go.mod和go.sum文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# 复制项目代码（按编译顺序复制，提高缓存命中率）
COPY internal/internal.go ./internal/
COPY internal/config/ ./internal/config/
COPY internal/infrastructure/ ./internal/infrastructure/
COPY internal/app/ ./internal/app/
COPY internal/main.go ./internal/

# 生成wire依赖注入代码
RUN wire ./internal/app

# 构建应用程序（使用静态链接和优化）
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" -o ixpay-pro ./internal/main.go

# 第二阶段：运行阶段
FROM alpine:3.19

# 更新包管理器并安装必要的工具
RUN apk update --no-cache && apk add --no-cache ca-certificates tzdata && rm -rf /var/cache/apk/*

# 创建非root用户运行应用
RUN adduser -D -g '' appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/ixpay-pro .

# 复制配置文件
COPY configs/config.yaml ./configs/

# 复制docs目录（包含Swagger文档）
COPY --from=builder /app/docs ./docs/

# 创建日志目录并设置权限
RUN mkdir -p /app/logs && chown -R appuser:appuser /app/logs && chmod 755 /app/logs

# 设置文件权限
RUN chmod +x ./ixpay-pro

# 切换到非root用户
USER appuser

# 声明暴露的端口
EXPOSE 8586

# 设置环境变量默认值
ENV ENV=production
LOG_LEVEL=info

# 设置健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8586/health || exit 1

# 设置启动命令
CMD ["./ixpay-pro"]