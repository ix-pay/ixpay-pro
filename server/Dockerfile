# 运行阶段（使用本地编译好的二进制文件）
FROM alpine:latest

WORKDIR /app

# 复制本地编译好的二进制文件
COPY cmd/build/ixpay-pro ./ixpay-pro

# 复制配置文件
COPY configs/config.yaml ./configs/

# 复制 docs 目录（包含 Swagger 文档）
COPY docs ./docs/

# 创建日志目录并设置权限
RUN mkdir -p /app/logs && chmod 755 /app/logs

# 设置文件权限
RUN chmod +x ./ixpay-pro

# 声明暴露的端口
EXPOSE 8081

# 设置环境变量默认值
ENV ENV=production \
    LOG_LEVEL=info

# 设置健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8081/health || exit 1

# 设置启动命令
CMD ["./ixpay-pro"]
