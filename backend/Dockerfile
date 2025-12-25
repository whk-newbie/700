# 多阶段构建
# 第一阶段：构建阶段
FROM golang:alpine AS builder

# 设置Go代理（使用国内镜像源）
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn

# 设置工作目录
WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git ca-certificates

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖（增加超时和重试）
RUN go mod download || go mod download

# 安装swag工具用于生成Swagger文档
RUN go install github.com/swaggo/swag/cmd/swag@latest

# 复制源代码
COPY . .

# 生成Swagger文档
RUN swag init -g cmd/server/main.go -o docs || echo "Warning: swag init failed, using existing docs"

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# 第二阶段：运行阶段
FROM alpine:latest

# 安装ca-certificates用于HTTPS请求
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 复制static目录（包含websocket文档和二维码等）
COPY --from=builder /app/static ./static

# 创建日志目录
RUN mkdir -p /root/logs

# 暴露端口
EXPOSE 8080

# 健康检查（使用wget或curl）
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health 2>/dev/null || exit 1

# 启动应用
CMD ["./main"]
