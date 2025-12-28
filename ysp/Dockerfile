# 构建阶段
FROM golang:1.21-alpine AS builder

ARG TARGETARCH
ARG TARGETVARIANT

WORKDIR /app

# 复制依赖文件并下载
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码并编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH GOARM=${TARGETVARIANT:1} go build -o ysp .

# 运行阶段（使用 scratch 最小镜像）
FROM scratch
COPY --from=builder /app/ysp /ysp
EXPOSE 9824
ENTRYPOINT ["/ysp"]
