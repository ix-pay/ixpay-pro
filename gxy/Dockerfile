# 第一阶段：构建阶段
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制项目依赖文件
COPY go.mod ./

# 下载依赖
RUN go mod download

# 复制项目源代码
COPY . .

# 构建Go应用程序
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gateway ./cmd/gateway

# 第二阶段：运行阶段
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 复制配置文件
COPY config.json ./

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/gateway ./

# 设置执行权限
RUN chmod +x ./gateway

# 暴露应用程序端口
EXPOSE 8385 8384

# 设置启动命令
CMD ["./gateway", "-config", "./config.json"]