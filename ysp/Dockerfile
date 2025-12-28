# 构建阶段
FROM golang:1.21-alpine AS builder

ARG TARGETARCH
ARG TARGETVARIANT

WORKDIR /app

# 复制所有文件
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH GOARM=${TARGETVARIANT:-0} \
    go build -o ysp .

# 运行阶段
FROM scratch
COPY --from=builder /app/ysp /ysp
EXPOSE 9824
ENTRYPOINT ["/ysp"]