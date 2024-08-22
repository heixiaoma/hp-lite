FROM golang:1.17-alpine

# 在容器内设置 /hello 为当前工作目录
WORKDIR /hp-pro-arm64
ADD hp-pro-arm64  /hp-pro-arm64/hp-pro-arm64
RUN chmod -R 777 /hp-pro-arm64

CMD ./hp-pro-arm64
#docker build -t hp-client .
