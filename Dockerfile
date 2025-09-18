# 构建阶段：编译 Go 可执行文件并拉取 qqwry.dat 内置到镜像
FROM golang:1.22-alpine AS builder

WORKDIR /workspace

# 构建时可覆盖数据来源
ARG QQWRY_URL="https://github.com/metowolf/qqwry.dat/releases/latest/download/qqwry.dat"

# 安装下载工具
RUN apk add --no-cache curl

COPY go.mod ./
RUN go mod download

COPY . .

# 在构建阶段下载 qqwry.dat（内置到镜像，避免运行时外网依赖）
RUN curl -L --fail --retry 3 --retry-delay 3 -o /workspace/qqwry.dat "$QQWRY_URL"

# 构建静态二进制
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ipservice .

# 运行阶段：使用 Alpine 作为基础镜像
FROM alpine:3.20

WORKDIR /app

# 运行阶段最小化：无需 CA 证书（由反向代理处理 TLS）

COPY --from=builder /workspace/ipservice ./ipservice
COPY --from=builder /workspace/qqwry.dat ./qqwry.dat
COPY docs ./docs

# 镜像内置 qqwry.dat，如需覆盖可通过挂载替换
ENV IP_API_QQWRY_PATH=/app/qqwry.dat
ENV IP_API_AUTO_FETCH=false
ENV IP_API_LISTEN=:8080

EXPOSE 8080

ENTRYPOINT ["/app/ipservice"]