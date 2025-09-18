# 构建阶段：编译 Go 可执行文件
FROM golang:1.22-alpine AS builder

WORKDIR /workspace

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ipservice ./cmd/server

# 运行阶段：使用极简镜像部署
FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /workspace/ipservice ./ipservice
COPY qqwry.dat ./qqwry.dat

ENV IP_API_QQWRY_PATH=/app/qqwry.dat
ENV IP_API_LISTEN=:8080

EXPOSE 8080

ENTRYPOINT ["/app/ipservice"]
