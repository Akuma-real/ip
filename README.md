# 基于 qqwry.dat 的 Go IP 信息查询服务

该项目提供一个使用 Go 实现的 RESTful API，可读取本地 `qqwry.dat` 库并返回 IP 归属信息，支持 Docker 容器化部署，适合在受限网络或离线环境中快速上线。

## 功能特性
- ⚡ 基于 Gin 框架的高性能 HTTP 服务，启动即加载数据，查询延迟低
- 🧵 内部采用内存映射与读写锁，满足多并发查询场景的线程安全需求
- 🔁 支持热加载（`Reload` 方法），便于后续扩展自动更新数据文件
- 🛠️ 通过环境变量灵活配置监听地址与数据文件路径

## 环境准备
1. 准备 `qqwry.dat` 文件，默认放置在项目根目录
2. 安装 Go 1.22+（若仅通过 Docker 构建，可无需本地安装）
3. 推荐执行 `go mod tidy` 自动生成 `go.sum`，确保依赖可复现

## 快速启动

### 使用 Docker 构建
```bash
docker build -t qqwry-ip-service .
docker run -d --name qqwry-ip-service \
  -p 8080:8080 \
  -v /path/to/qqwry.dat:/app/qqwry.dat:ro \
  qqwry-ip-service
```
容器启动后，可通过 `http://localhost:8080/health` 进行探活检查。

### 本地运行（Go 环境）
```bash
go run ./cmd/server
```
如需自定义监听地址或数据路径，可使用环境变量：
```bash
set IP_API_LISTEN=:9000
set IP_API_QQWRY_PATH=D:\\data\\qqwry.dat
```

## API 设计
- `GET /health`：返回 `{ "status": "ok" }` 用于健康检查
- `GET /ip/{ip}`：通过路径参数查询某个 IPv4 的归属信息
- `POST /ip`：请求体 `{"ip": "8.8.8.8"}`，适合与其他系统集成

响应示例：
```json
{
  "ip": "8.8.8.8",
  "country": "美国",
  "area": "谷歌公司",
  "raw": ["美国", "谷歌公司"]
}
```

## 目录结构
```
cmd/server/           # 程序入口，加载配置并启动 Web 服务
internal/config/      # 配置读取与校验逻辑
internal/ipdb/        # qqwry 数据解析与查询实现
internal/server/      # Gin 路由与请求处理
legacy/               # 旧版 Python 实现备份，可按需参考
qqwry.dat             # IP 数据库文件（需单独下载）
Dockerfile            # 多阶段构建镜像
```

## 扩展与优化建议
- 定期更新 `qqwry.dat`（可结合定时任务与 `Service.Reload`）
- 集成 Prometheus 指标或结构化日志，提升可观测性
- 增加 LRU 缓存以优化热点 IP 查询的延迟
- 对接 API 网关或认证模块，强化安全与访问控制

> 注：若仍需 Python 版本，可参考 `legacy/` 目录中的 FastAPI 实现。两种方案可根据业务需求任选其一。