FROM golang:1.17-alpine

# 在容器内设置 /hello 为当前工作目录
WORKDIR /hp-lite-amd64
ADD hp-lite-amd64  /hp-lite-amd64/hp-pro-amd64
RUN chmod -R 777 /hp-lite-amd64

CMD ./hp-lite-amd64
#docker build -t hp-client .
