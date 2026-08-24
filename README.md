# Project 07: Geofence Event Service

纯Go地理围栏事件处理与位置策略服务。项目提供终端注册、圆形/多边形围栏、位置点接收、进入/离开/停留事件、事件状态操作以及带HMAC签名的Webhook订阅。默认使用线程安全内存仓储，服务无需外部数据库即可启动；`migrations/` 提供PostgreSQL/PostGIS生产表结构，仓储与发送器均通过接口隔离，便于接入生产实现。

## 目录

- `cmd/geofenced`：进程入口、优雅关闭和HTTP Server配置。
- `internal/domain`：terminal、geofence、location、event、subscription领域模型及规则。
- `internal/application`：用例服务、仓储端口、位置到事件的编排。
- `internal/adapter/http`：REST API、JSON校验、统一错误和请求指标。
- `internal/infrastructure`：内存仓储、时钟、结构化日志、Prometheus文本指标、Webhook客户端。
- `api`：OpenAPI概要；`configs`：配置示例；`migrations`：PostGIS迁移；`deploy`：Docker文件；`scripts`：本地运行和检查。

## 启动

需要Go 1.22或更新版本：

```bash
go run ./cmd/geofenced
```

默认监听 `:8087`。环境变量 `HTTP_ADDR`、`LOG_LEVEL`、`DWELL_SECONDS`、`SHUTDOWN_SECONDS` 可覆盖配置。按Ctrl-C或发送SIGTERM会停止接收新请求并等待活动连接结束。

浏览器访问 `http://localhost:8087/console/` 可打开运维控制台，查看服务状态、终端、围栏和最近事件。

## 核心流程示例

```bash
curl -s localhost:8087/healthz
curl -s -X POST localhost:8087/api/v1/terminals -H 'Content-Type: application/json' \
  -d '{"id":"truck-01","name":"Truck 01","tags":{"site":"north"}}'
curl -s -X POST localhost:8087/api/v1/geofences -H 'Content-Type: application/json' \
  -d '{"id":"yard","name":"North Yard","type":"circle","center":{"latitude":35.6812,"longitude":139.7671},"radius_m":500}'
curl -s -X POST localhost:8087/api/v1/locations -H 'Content-Type: application/json' \
  -d '{"terminal_id":"truck-01","latitude":35.6812,"longitude":139.7671,"observed_at":"2026-08-22T00:00:00Z"}'
curl -s 'localhost:8087/api/v1/events?terminal_id=truck-01'
```

首次位置上报若在围栏内，会返回`enter`事件；之后在围栏外上报会返回`exit`事件。连续在围栏内达到`DWELL_SECONDS`后会返回一次`dwell`事件。位置点按`observed_at`排序，重复事件通过幂等键去重。地理坐标使用WGS84，经纬度范围会严格校验。

## API摘要

- `POST/GET /api/v1/terminals`、`GET/PATCH /api/v1/terminals/{id}`：终端及标签、状态。
- `POST/GET /api/v1/geofences`、`GET/PATCH /api/v1/geofences/{id}`：围栏几何、版本和active/paused模式。
- `POST /api/v1/locations`：上报位置并同步返回新事件。
- `POST /api/v1/replay`：按观测时间重放终端位置历史，并提交一致性 checkpoint。
- `GET /api/v1/terminals/{id}/locations`：查询位置历史。
- `GET /api/v1/events`、`GET /api/v1/events/{id}`、`POST /api/v1/events/{id}/ack|close`：事件筛选和生命周期。
- `POST/GET /api/v1/subscriptions`、`GET /api/v1/subscriptions/{id}/deliveries`：Webhook订阅和投递记录。
- `GET /healthz`、`GET /readyz`、`GET /metrics`：运行状态和指标。
- `GET /api/v1/operations/summary`：控制台使用的原子计数摘要。

Webhook请求体包含`event`和`sent_at`，配置secret时会附带`X-Geofence-Signature`（HMAC-SHA256十六进制）。非2xx响应会记录失败投递，生产部署可在`WebhookSender`接口外包裹队列与重试策略。

## 验证

```bash
./scripts/check.sh
```

该脚本执行gofmt、`go test ./...`并检查至少2000行非测试Go源码。迁移脚本需要PostgreSQL和PostGIS；默认内存模式用于本地功能验证。
