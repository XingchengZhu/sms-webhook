# 使用通用的基础镜像
FROM 10.29.230.150:31381/library/golang:1.19-alpine AS builder

# 设置工作目录
WORKDIR /app

# 拷贝 go.mod 和 go.sum 并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 拷贝源码并编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -o /app/sms-webhook

# 创建最终运行镜像
FROM 10.29.230.150:31381/library/golang:1.19-alpine

# 设置工作目录
WORKDIR /root/

# 拷贝编译后的二进制文件
COPY --from=builder /app/sms-webhook .

# 暴露端口
EXPOSE 8080

# 启动服务
CMD ["./sms-webhook"]
