FROM golang:1.17-alpine

# 在容器内设置 /hello 为当前工作目录
WORKDIR /hp-client-golang-amd64
ADD hp-client-golang-amd64  /hp-client-golang-amd64/hp-client-golang-amd64
RUN chmod -R 777 /hp-client-golang-amd64

EXPOSE 10240

CMD ./hp-client-golang-amd64

#docker build -t hp-client .
