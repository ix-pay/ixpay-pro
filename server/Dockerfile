# 第一阶段：构建阶段
FROM golang:1.24.6-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装构建依赖\ RUN apk add --no-cache git gcc musl-dev

# 复制go.mod和go.sum文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制整个项目代码
COPY . .

# 生成wire依赖注入代码
RUN go run github.com/google/wire/cmd/wire ./internal/app

# 构建应用程序
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ixpay-pro .

# 第二阶段：运行阶段
FROM alpine:3.19

# 创建非root用户运行应用
RUN adduser -D -g '' appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/ixpay-pro .

# 复制配置文件
COPY configs/config.yaml ./configs/

# 创建日志目录
RUN mkdir -p /app/logs && chown -R appuser:appuser /app/logs

# 切换到非root用户
USER appuser

# 声明暴露的端口
EXPOSE 8586

# 设置启动命令
CMD ["./ixpay-pro"]