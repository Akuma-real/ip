# Repository Guidelines

## 项目结构与模块组织
- `cmd/server`：Go 程序入口，加载配置并启动 HTTP 服务。
- `internal/config`：负责读取环境变量、校验并补全 `qqwry.dat` 绝对路径。
- `internal/ipdb`：封装 QQWry 数据解析、查询与热加载逻辑，是业务核心。
- `internal/server`：基于 Gin 的路由与处理器集合，统一注入 `ipdb.Service`。
- `docs/`：存放数据格式与协议说明，可扩展更多运维文档。
- 根目录包含 `Dockerfile` 与 `qqwry.dat`（运行时挂载），其余生成物请保持忽略。

## 构建、测试与开发命令
- `go run ./cmd/server`：本地快速启动，读取当前目录下的 `qqwry.dat`。
- `go build ./cmd/server`：产出自包含二进制，适合部署到裸机或容器。
- `go test ./... -cover`：运行全部包的单元测试并输出覆盖率，提交前必跑。
- `docker build -t qqwry-ip-service .` 与 `docker run -p 8080:8080 -v <host>/qqwry.dat:/app/qqwry.dat:ro qqwry-ip-service`：完成容器化验证。

## 代码风格与命名约定
- 所有 Go 代码执行 `gofmt`，默认使用 Tab 缩进与驼峰式导出命名。
- 包名保持短小且全小写，文件名用下划线分隔，如 `router.go`、`service.go`。
- 结构体字段与 JSON 输出需显式声明 Tag；日志与错误信息统一中文语境并提供上下文。

## 测试指南
- 测试文件放置于同包目录，命名 `xxx_test.go`，函数以 `Test`、`Benchmark` 前缀。
- 针对 `ipdb.Service` 重点覆盖：初始化失败场景、边界 IP 查询、热加载并发安全。
- 结合 `go test ./internal/ipdb -race` 检查数据竞态，保证读写锁逻辑正确。

## 提交与拉取请求指南
- 当前历史仅有 `Initial commit`，后续请使用祈使句或 Conventional Commits（如 `feat: add ip range cache`）。
- 每次提交前执行格式化与测试，必要时附带 `go mod tidy` 结果。
- 拉取请求需包含：变更摘要、验证步骤（命令输出或截图）、相关 Issue 链接与回滚评估。

## 安全与配置提示
- 保持 `qqwry.dat` 不进入版本控制，部署时通过环境变量 `IP_API_QQWRY_PATH` 指向绝对路径。
- 暴露端口默认 `:8080`，在受限网络中请结合反向代理或安全组限制来源 IP。
- 若启用 Docker，确保宿主机数据卷只读挂载，防止意外覆盖数据库文件。