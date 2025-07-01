FROM alpine:latest
ARG TARGETARCH
ARG TARGETVARIANT
WORKDIR /hp-lite
# 根据架构选择正确的二进制文件
COPY --chmod=777 hp-lite-${TARGETARCH}${TARGETVARIANT} /hp-lite/hp-lite
CMD ["./hp-lite"]