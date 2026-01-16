# 使用本地已编译的Linux二进制文件（包含嵌入的前端）
FROM alpine:latest
WORKDIR /hp-lite-server

# 安装必要的运行时依赖
RUN apk add --no-cache ca-certificates tzdata libc6-compat

# 复制已编译的Linux二进制文件（包含嵌入的最新前端）
COPY hp-server-golang/target/hp-lite-server-amd64 /hp-lite-server/hp-server

# 创建配置文件目录并设置权限
RUN mkdir -p /hp-lite-server/config && \
    chmod +x /hp-lite-server/hp-server

# 暴露端口
EXPOSE 9999 16666 9091/tcp 9091/udp

# 设置时区
ENV TZ=Asia/Shanghai

# 启动命令
CMD ["./hp-server"]
