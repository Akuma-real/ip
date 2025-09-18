# 基于 qqwry.dat 的 Go IP 信息查询服务

该项目提供一个使用 Go 实现的 RESTful API，可读取本地 `qqwry.dat` 库并返回 IP 归属信息，支持 Docker 容器化部署，适合在受限网络或离线环境中快速上线。

## 功能特性
- ⚡ 基于 Gin 框架的高性能 HTTP 服务，启动即加载数据，查询延迟低
- 🧵 内部采用内存映射与读写锁，满足多并发查询场景的线程安全需求
- 🔁 支持热加载（`Reload` 方法），便于后续扩展自动更新数据文件
- 🛠️ 通过环境变量灵活配置监听地址与数据文件路径

## 环境准备
1. 数据文件 `qqwry.dat`
   - 项目支持自动获取：默认在启动时若检测到本地缺失，将从以下地址下载到 `IP_API_QQWRY_PATH` 指定位置：
     https://github.com/metowolf/qqwry.dat/releases/latest/download/qqwry.dat
   - 可配置项：
     - `IP_API_QQWRY_PATH`（默认 `qqwry.dat`，相对路径将自动转绝对路径）
     - `IP_API_QQWRY_URL`（默认上述地址）
     - `IP_API_AUTO_FETCH`（默认 `true`，可设为 `false` 关闭自动下载）
   - 也可手动下载：
     - Windows PowerShell: `Invoke-WebRequest -Uri https://github.com/metowolf/qqwry.dat/releases/latest/download/qqwry.dat -OutFile .\qqwry.dat`
     - Linux/macOS: `curl -L -o qqwry.dat https://github.com/metowolf/qqwry.dat/releases/latest/download/qqwry.dat`
   - 容器运行：未挂载文件时，将尝试写入 `IP_API_QQWRY_PATH` 路径；如使用只读挂载，请提前准备数据文件。
2. 安装 Go 1.22+（若仅通过 Docker 构建，可无需本地安装）
3. 推荐执行 `go mod tidy` 自动生成 `go.sum`，确保依赖可复现

## 快速启动

### 使用 Docker 构建
```bash
docker build -t qqwry-ip-service .
docker run -d --name qqwry-ip-service \
  -p 8080:8080 \
  qqwry-ip-service
```
提示：镜像在构建阶段已从上游拉取并内置 `qqwry.dat`。如需指定数据版本，可：
- 构建时覆盖下载地址：`docker build --build-arg QQWRY_URL=https://your.mirror/qqwry.dat -t qqwry-ip-service .`
- 运行时覆盖为宿主机文件：`-v /path/to/qqwry.dat:/app/qqwry.dat:ro -e IP_API_QQWRY_PATH=/app/qqwry.dat`
说明：镜像在构建阶段已内置 qqwry.dat（可通过构建参数或挂载覆盖）。
容器启动后，可通过 `http://localhost:8080/health` 进行探活检查。

### 本地运行（Go 环境）
```bash
go run .
```
如需自定义监听地址或数据路径，可使用环境变量：
```bash
set IP_API_LISTEN=:9000
set IP_API_QQWRY_PATH=D:\\data\\qqwry.dat
```

## API 设计
- `GET /`：动态文档页（基于当前访问域名/协议生成可点击链接与 curl 示例，支持在线试用）
- `GET /docs`：API 使用说明（docs/api_usage.md 渲染）
- `GET /health`：返回 `{ "status": "ok" }` 用于健康检查
- `GET /ip`：直接返回当前访问者的 IP 归属信息，自动识别 `X-Forwarded-For` 等代理头
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

### 直接访问示例
- 浏览器测试：访问 `http://localhost:8080/ip` 可直接获取当前客户端的归属信息
- 如果出现 `当前qqwry.dat仅支持IPv4查询`，请改用 `http://127.0.0.1:8080/ip` 或 `curl --ipv4 http://localhost:8080/ip` 强制使用 IPv4 连接
- 单次查询：`curl http://localhost:8080/ip/8.8.8.8`
- 动态识别客户端 IP：`curl http://localhost:8080/ip -H "X-Forwarded-For: 1.2.3.4"`
- JSON 集成：`curl -X POST http://localhost:8080/ip -d '{"ip":"8.8.8.8"}' -H "Content-Type: application/json"`

## 目录结构
```
main.go               # 程序入口，加载配置并启动 Web 服务（含优雅关停与超时配置）
internal/config/      # 配置读取与校验逻辑
internal/ipdb/        # qqwry 数据解析与查询实现（含领域错误）
internal/server/      # Gin 路由与请求处理
qqwry.dat             # IP 数据库文件（不纳入版本控制；构建时内置/运行时可挂载覆盖）
Dockerfile            # 多阶段构建镜像（构建时拉取并内置数据文件）
```

## 扩展与优化建议
- 定期更新 `qqwry.dat`（可结合定时任务与 `Service.Reload`）
- 集成 Prometheus 指标或结构化日志，提升可观测性
- 增加 LRU 缓存以优化热点 IP 查询的延迟
- 对接 API 网关或认证模块，强化安全与访问控制

