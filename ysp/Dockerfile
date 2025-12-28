# 构建阶段
FROM golang:1.21-alpine AS builder

ARG TARGETARCH
ARG TARGETVARIANT

WORKDIR /app

# 复制go模块文件
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源代码
COPY . .

# 构建应用（禁用CGO，静态链接）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH GOARM=${TARGETVARIANT:-0} \
    go build -ldflags="-s -w" -o ysp .

# 运行阶段（使用scratch最小镜像）
FROM scratch
COPY --from=builder /app/ysp /ysp
EXPOSE 9824
ENTRYPOINT ["/ysp"]
