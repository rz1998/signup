# 报名系统后端 Dockerfile

FROM golang:1.22-alpine AS builder

# 设置Go模块代理（国内镜像）
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on
ENV GOSUMDB=off

# 安装构建依赖
RUN apk add --no-cache git make

WORKDIR /app

# 复制源代码
COPY . .

# 下载依赖
RUN go mod tidy

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o signup-api .

# 运行镜像
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 复制构建产物
COPY --from=builder /app/signup-api .
COPY --from=builder /app/etc ./etc

# 创建非root用户
RUN adduser -D -u 1000 appuser
USER appuser

# 暴露端口
EXPOSE 8082

# 启动命令
CMD ["./signup-api"]
