# 分布式游戏排行榜系统 Dockerfile
# Author: HHaou
# Created: 2024-01-20

# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o ranking-server \
    ./cmd/server

# 运行阶段
FROM alpine:latest

# 安装ca证书和时区数据
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN addgroup -g 1001 -S ranking && \
    adduser -u 1001 -S ranking -G ranking

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/ranking-server .

# 复制配置文件
COPY --from=builder /app/configs ./configs

# 创建日志目录
RUN mkdir -p /app/logs && \
    chown -R ranking:ranking /app

# 切换到非root用户
USER ranking

# 暴露端口
EXPOSE 8080 9090

# 健康检查
HEALTHCHEK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动应用
CMD ["./ranking-server", "--config", "configs/config.yaml"]