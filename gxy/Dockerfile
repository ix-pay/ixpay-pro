# 构建阶段
FROM golang:1.26.5-alpine AS builder

WORKDIR /build

# 安装 git（go mod 需要）
RUN apk add --no-cache git

# 复制 go.mod 和 go.sum
COPY go.mod ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/

# 构建二进制文件
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cmd/build/gateway ./cmd/gateway/main.go

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 安装 ca 证书（用于 https 请求）
RUN apk --no-cache add ca-certificates

# 复制构建好的二进制文件
COPY --from=builder /build/cmd/build/gateway ./gateway

RUN chmod +x ./gateway

EXPOSE 8085 8084
