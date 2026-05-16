# 运行阶段（使用本地编译好的二进制文件）
FROM alpine:latest

WORKDIR /app

# 复制本地编译好的二进制文件
COPY cmd/build/gateway ./gateway

RUN chmod +x ./gateway

EXPOSE 8385 8384
