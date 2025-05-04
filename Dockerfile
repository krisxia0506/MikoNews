FROM golang:1.24-alpine AS builder

WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o miko_news ./cmd/main.go

# 使用小体积的alpine镜像
FROM alpine:latest

# 安装ca-certificates，确保HTTPS工作正常
RUN apk --no-cache add ca-certificates tzdata

# 设置时区为Asia/Shanghai
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从builder阶段复制编译好的可执行文件
COPY --from=builder /app/miko_news .

# 创建configs目录并复制配置文件
COPY --from=builder /app/configs /app/configs

# 创建migrations目录并复制SQL文件
COPY --from=builder /app/migrations /app/migrations

# 暴露端口 (默认8080，但可通过配置文件修改)
EXPOSE 8080

# 运行服务
CMD ["./miko_news"] 