# 构建阶段
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

ARG TARGETARCH
ARG TARGETVARIANT
ARG BUILDPLATFORM

WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 设置构建参数
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH GOARM=${TARGETVARIANT:-0} \
    go build -ldflags="-s -w" -o ysp .

# 运行阶段（最小化镜像）
FROM scratch
COPY --from=builder /app/ysp /ysp
EXPOSE 9824
ENTRYPOINT ["/ysp"]
