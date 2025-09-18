# API 使用说明
本说明整理服务支持的 HTTP 接口以及直接访问场景下的最佳实践。

## 接口列表
- `GET /health`：健康探针，返回 `{ "status": "ok" }`。
- `GET /ip`：返回当前访问者的 IP 归属信息，会综合 `X-Forwarded-For`、`X-Real-IP` 与连接源地址。
- `GET /ip/{ip}`：根据路径参数查询指定 IPv4。
- `POST /ip`：接收 `{ "ip":"8.8.8.8" }` 形式的 JSON 请求体。

## 客户端 IP 判定规则
- 优先读取 `X-Forwarded-For` 的首个合法 IPv4，其次为 `X-Real-IP`。
- 若未包含代理头，则回退为真实连接地址。
- 如无法识别合法 IPv4，将返回 `400` 并提示“无法识别客户端IP”。

## 快速体验示例
- 浏览器直接访问 `http://localhost:8080/ip`，即可验证自身出口地址。
- 若提示 `当前qqwry.dat仅支持IPv4查询`，请改用 `http://127.0.0.1:8080/ip` 或在 curl 中追加 `--ipv4`，强制使用 IPv4 连接
- 指定查询目标：`curl http://localhost:8080/ip/8.8.8.8`
- 代理场景模拟：`curl http://localhost:8080/ip -H "X-Forwarded-For: 1.2.3.4"`
- JSON 集成：`curl -X POST http://localhost:8080/ip -H "Content-Type: application/json" -d '{"ip":"8.8.8.8"}'`

## 响应字段说明
- `ip`：最终确认的查询目标地址。
- `country`：归属国家/地区，若未知则为空字符串。
- `area`：归属运营商或网络区域，若未知则为空字符串。
- `raw`：原始字段数组，便于保留未经归一化的描述。
