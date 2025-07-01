FROM alpine:latest
ARG TARGETARCH
ARG TARGETVARIANT
WORKDIR /hp-lite-server
COPY --chmod=777 hp-lite-server-${TARGETARCH}${TARGETVARIANT}  /hp-lite-server/hp-lite-server
CMD ["./hp-lite-server"]