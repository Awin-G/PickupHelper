# PickupHelper 后端开发文档

> 快递代取与驿站管理系统 — Go/Gin 后端

---

## 1. 项目概览

### 1.1 仓库结构

```
PickupHelper/                         # 仓库根目录
├── backend/                          # 后端 Go 项目（本文档范围）
│   ├── cmd/server/                   # 入口 main.go + 手动 DI 容器
│   ├── internal/                     # 业务代码（不对外暴露）
│   │   ├── config/                   # viper 配置加载
│   │   ├── errors/                   # 统一错误码 + AppError + HTTP 映射
│   │   ├── handler/                  # HTTP 处理器（12 个注册函数，34 个端点）
│   │   ├── log/                      # slog 封装 + trace_id 注入
│   │   ├── middleware/               # JWT / CORS / RateLimit / Recovery / Trace / Validator / AdminOnly
│   │   ├── models/                   # 领域模型 + DTO + 常量 + 工具函数
│   │   ├── repository/               # 数据访问层（sqlx 原生 SQL）+ DBTX 接口 + WithTx
│   │   ├── router/                   # 路由注册 + 中间件链编排
│   │   ├── server/                   # http.Server 封装（优雅关闭）
│   │   └── service/                  # 业务逻辑层（10 个 service）
│   ├── test/                         # 集成测试（testcontainers MySQL+Redis）+ fixtures
│   ├── docs/                         # 本文档目录
│   ├── configs/                      # config.dev.yaml / config.test.yaml
│   ├── migrations/                   # goose SQL 迁移脚本
│   ├── Makefile                      # 构建/测试/迁移命令
│   └── go.mod / go.sum               # 模块名 pickup-helper，Go 1.26.4
├── 需求规格说明/                      # 需求文档
├── 详细设计文档/                      # 详细设计 + DDL + API 契约（黄金标准）
└── CONTEXT-backend.md                # 后端开发上下文入口
```

### 1.2 技术栈

| 维度 | 选型 | 版本 |
|------|------|------|
| 语言 | Go | 1.26.4 |
| Web 框架 | Gin | v1.12 |
| 数据库 | MySQL 8.0 (MariaDB 兼容) | — |
| SQL 库 | sqlx (原生 SQL，不用 ORM) | v1.4 |
| 缓存 | Redis | 7+ |
| 迁移 | goose | v3 |
| 配置 | viper (YAML + 环境变量覆盖) | v1.21 |
| 日志 | log/slog (结构体 JSON) | stdlib |
| 校验 | go-playground/validator v10 | v10 |
| 鉴权 | golang-jwt/jwt v5 (access + refresh 双 token) | v5 |
| 密码 | bcrypt | x/crypto |
| 测试 | testify + testcontainers (MySQL+Redis 真容器) | — |
| 架构 | 单体模块化，三层 handler→service→repository | — |

### 1.3 已实现模块（8 个 Phase）

| Phase | 模块 | 端点数 | 测试数 | 核心功能 |
|-------|------|--------|--------|----------|
| 1 | Foundation | — | — | 骨架/配置/日志/迁移/中间件/JWT/测试基建 |
| 2 | User | 8 | 14 | 验证码登录/注册/用户信息/跑腿员审核/黑名单 |
| 3 | Parcel | 6 | 15 | 包裹入库/取件码生成/货架分配/批量导入/列表/状态变更 |
| 4 | Pickup | 4 | 12 | 核销取件/自助出库/扫码批量出库/取件日志 |
| 5 | Proxy | 7 | 18 | 发布代取/任务大厅/接单/送达/确认/取消/我的订单 |
| 6 | Shelf | 4 | 9 | 货架 CRUD/容量管理/占用热力图 |
| 7 | Notification | 2 | 5 | 通知列表/批量已读/未读计数 |
| 8 | Stats | 3 | 6 | 首页看板/包裹趋势/代取财务汇总 |

**合计：34 个 API 端点，84 个集成测试，86 个 Go 源文件。**

---

## 2. 开发环境

### 2.1 系统要求

| 依赖 | 最低版本 | 用途 |
|------|----------|------|
| Go | 1.26+ | 编译运行 |
| Docker | 24+ | 本地 MySQL/Redis 容器 + 集成测试 |
| mysql 客户端 | 8.0 | 数据库迁移（`make migrate-up`） |
| goose | v3 | 数据库迁移（`go install github.com/pressly/goose/v3/cmd/goose@latest`） |

### 2.2 快速启动

```bash
# 1. 进入后端目录
cd backend

# 2. 安装 goose（如未安装）
go install github.com/pressly/goose/v3/cmd/goose@latest

# 3. 启动 MySQL + Redis 容器
docker run -d --name pickup-mysql -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=1973 \
  -e MYSQL_DATABASE=pickup_helper \
  mysql:8.0

docker run -d --name pickup-redis -p 6379:6379 redis:7-alpine

# 4. 数据库迁移
make migrate-up

# 5. 启动服务（开发/测试时使用 18080 端口，避免与前端 8080 冲突）
APP_PORT=18080 make run
```

### 2.3 常用命令

```bash
# 编译
make build                          # → bin/server

# 运行（dev 配置 + 可指定端口）
APP_PORT=18080 make run             # 推荐：开发端口
make run                            # 默认 :8080

# 测试
make test-unit                      # 单元测试（internal/ 下所有包）
make test-integration               # 集成测试（test/ 下所有包，需 Docker）
make vet                            # 静态检查
make vet-integration                # 集成测试代码静态检查

# 数据库
make migrate-up                     # 执行所有待执行迁移
make migrate-down                   # 回滚最近一次迁移
make migrate-reset                  # 重置（回滚所有 + 重新执行）

# 代码格式化
make fmt                            # gofmt -s -w
make tidy                           # go mod tidy
```

### 2.4 端口约定

| 端口 | 用途 | 使用者 |
|------|------|--------|
| **8080** | 生产/前端 agent 默认端口 | 前端 agent |
| **18080** | 后端开发/测试端口 | 后端 agent |
| **3306** | MySQL | 本地容器 |
| **6379** | Redis | 本地容器 |

**后端开发和测试时永不占用 8080。**

---

## 3. 部署环境

### 3.1 生产环境要求

| 组件 | 要求 | 说明 |
|------|------|------|
| Go | 1.26+ | 服务运行 |
| MySQL | 8.0 / MariaDB 10.5+ | 字符集 utf8mb4，时区 Asia/Shanghai |
| Redis | 7+ | 缓存（验证码/限流）+ 分布式锁（定时任务） |
| Docker | 可选 | 容器化部署 |
| 端口 | 8080 | 默认监听端口 |
| HTTPS | 推荐 | 生产环境强制 |
| 域名 | 视部署而定 | CORS 白名单配置 |

### 3.2 生产部署步骤

```bash
# 1. 编译二进制
cd backend
make build                    # → bin/server

# 2. 准备配置文件
cp configs/config.dev.yaml configs/config.prod.yaml
vim configs/config.prod.yaml  # 修改 MySQL/Redis/JWT/CORS 为生产值

# 3. 数据库迁移
APP_ENV=prod goose -dir migrations mysql "<DSN>" up

# 4. 启动服务
APP_ENV=prod ./bin/server
```

### 3.3 环境变量覆盖

配置中所有字段均可通过环境变量覆盖（viper AutomaticEnv + `__` 分隔符）：

| 环境变量 | 覆盖配置项 | 示例 |
|----------|-----------|------|
| `APP_ENV` | 配置文件选择 | `prod` → `configs/config.prod.yaml` |
| `APP_PORT` | 服务端口（Makefile 会转为 `SERVER__PORT`） | `APP_PORT=18080 make run` |
| `SERVER__PORT` | `server.port` | `SERVER__PORT=9090` |
| `MYSQL__HOST` | `mysql.host` | `MYSQL__HOST=10.0.0.1` |
| `MYSQL__PASSWORD` | `mysql.password` | `MYSQL__PASSWORD=prod-secret` |
| `REDIS__HOST` | `redis.host` | `REDIS__HOST=redis.internal` |
| `JWT__ACCESS_SECRET` | `jwt.access_secret` | `JWT__ACCESS_SECRET=<random>` |
| `JWT__REFRESH_SECRET` | `jwt.refresh_secret` | `JWT__REFRESH_SECRET=<random>` |
| `LOG__LEVEL` | `log.level` | `LOG__LEVEL=info` |
| `CORS__ALLOWED_ORIGINS` | `cors.allowed_origins`（逗号分隔） | `CORS__ALLOWED_ORIGINS=https://app.com` |

### 3.4 Docker 部署（推荐）

```dockerfile
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o bin/server ./cmd/server

FROM alpine:3
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai
COPY --from=builder /app/bin/server /usr/local/bin/server
COPY --from=builder /app/configs /configs
COPY --from=builder /app/migrations /migrations
EXPOSE 8080
CMD ["server"]
```

---

## 4. 配置详解

配置文件路径：`backend/configs/config.<env>.yaml`，`<env>` 由 `APP_ENV` 决定（默认 `dev`）。

### 4.1 server

```yaml
server:
  port: 8080              # 监听端口（dev 用 8080，test 用 8081）
  read_timeout: 15s       # 读取请求超时
  write_timeout: 15s      # 写入响应超时
```

### 4.2 mysql

```yaml
mysql:
  host: 127.0.0.1
  port: 3306
  username: root
  password: "1973"
  database: pickup_helper
  max_open_conns: 50      # 最大连接数
  max_idle_conns: 10      # 最大空闲连接数
  conn_max_lifetime: 300s # 连接最大存活时间
```

DSN 自动生成（含 `parseTime=true&loc=Asia%2FShanghai&charset=utf8mb4`）。

### 4.3 redis

```yaml
redis:
  host: 127.0.0.1
  port: 6379
  password: ""
  db: 0
  pool_size: 50           # 连接池大小
```

### 4.4 jwt

```yaml
jwt:
  access_secret: "dev-access-secret-change-me"      # 生产必须修改
  refresh_secret: "dev-refresh-secret-change-me"    # 生产必须修改
  access_ttl: 2h          # Access Token 2 小时
  refresh_ttl: 168h       # Refresh Token 7 天
  issuer: pickup-helper
```

### 4.5 log

```yaml
log:
  level: debug             # debug | info | warn | error
  format: json             # json | text
```

### 4.6 rate_limit

```yaml
rate_limit:
  qps: 100                 # 每秒令牌数（0 = 禁用）
  burst: 200               # 突发容量
```

### 4.7 cors

```yaml
cors:
  allowed_origins:
    - "http://localhost:5173"   # Taro 小程序开发
    - "http://localhost:3000"   # PureAdmin 管理端
```

---

## 5. 架构设计

### 5.1 三层架构

```
handler (Gin)  →  service (业务逻辑)  →  repository (SQL)
     ↑                 ↑                     ↑
  DTO/请求响应      领域逻辑/事务         DBTX 接口
```

### 5.2 中间件链

```
Recovery → TraceID → Logger → CORS → RateLimit → JWT → AdminOnly
```

### 5.3 鉴权体系

- **Access Token**：JWT HS256，2h 有效期，携带在 `Authorization: Bearer <token>`
- **Refresh Token**：JWT HS256，7d 有效期，用于 `/auth/refresh` 换新 access token
- **Claims 包含**：`user_id`, `user_type`, `station_id`, `role`
- **JWT 中间件**：解析后向内覆盖 `X-User-Id`/`X-User-Type`/`X-Station-Id`/`X-Role` 请求头，防止客户端伪造
- **AdminOnly 中间件**：检查 `X-Role=admin`，非 admin 返回 403

### 5.4 事务管理

```go
err := repository.WithTx(ctx, db, func(tx *sqlx.Tx) error {
    // 多次 repo 调用，tx 满足 DBTX 接口
    if e := repoA.Create(ctx, tx, ...); e != nil { return e }
    if e := repoB.Update(ctx, tx, ...); e != nil { return e }
    return nil
})
// 自动 commit（成功）或 rollback（panic/error）
```

### 5.5 统一响应

```json
{"code": 0, "msg": "success", "data": {...}, "trace_id": "a1b2c3d4e5f6"}
```

- `code=0` 成功，非 0 业务错误
- HTTP 状态码由 `code` 通过 `HTTPStatus()` 函数映射
- 错误码分段：10xxx 通用，101xx User，102xx Parcel，103xx Pickup，104xx Proxy，105xx Shelf，106xx Notify，107xx Station

### 5.6 错误码速查

| 范围 | 模块 | 示例 |
|------|------|------|
| 0 | 成功 | — |
| 10001-10010 | 通用 | 参数错误/未登录/权限不足/内部错误 |
| 10101-10171 | User | 手机格式/验证码错误/黑名单/审核 |
| 10201-10243 | Parcel | 重复入库/货架已满/包裹不存在/状态非法 |
| 10301-10313 | Pickup | 取件码无效/已取件/位置异常 |
| 10401-10443 | Proxy | 非本人包裹/接单/送达/确认/取消 |
| 10501-10513 | Shelf | 编号冲突/容量非法/不存在 |
| 10601 | Notify | 通知不存在 |

完整定义见 `internal/errors/codes.go`。

---

## 6. API 端点

### 6.1 公开端点（无 JWT）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/auth/send-code` | 发送短信验证码 |
| POST | `/auth/login` | 手机号验证码登录/注册 |
| POST | `/auth/refresh` | refresh token 换新 access token |
| POST | `/admin/auth/login` | 管理员用户名密码登录 |
| GET | `/health` | 存活检查 |
| GET | `/health/ready` | 就绪检查（MySQL + Redis） |

### 6.2 用户端点（JWT）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/user/info` | 获取用户信息 |
| PUT | `/user/info` | 更新用户信息 |
| POST | `/user/runner/apply` | 申请跑腿员 |
| GET | `/parcels/my` | 我的包裹 |
| GET | `/parcels/:id/pickup-code` | 获取取件码 |
| POST | `/pickup/self-checkout` | 自助出库 |
| POST | `/pickup/scan-station` | 扫码批量出库 |
| POST | `/proxy/publish` | 发布代取 |
| GET | `/proxy/tasks` | 任务大厅 |
| POST | `/proxy/accept/:id` | 接单 |
| POST | `/proxy/request-delivery/:id` | 发起送达确认 |
| POST | `/proxy/confirm-delivery/:id` | 确认收货 |
| POST | `/proxy/cancel` | 取消订单 |
| GET | `/proxy/my-orders` | 我的代取订单 |
| GET | `/notifications` | 通知列表 |
| PUT | `/notifications/read` | 标记已读 |

### 6.3 管理员端点（JWT + AdminOnly）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/admin/user/runner/applications` | 跑腿员申请列表 |
| PUT | `/admin/user/runner/applications/:id/audit` | 审核跑腿员申请 |
| PUT | `/admin/users/:id/blacklist` | 黑名单管理 |
| POST | `/parcels/scan-in` | 包裹入库 |
| GET | `/parcels` | 包裹列表 |
| PUT | `/parcels/:id/status` | 变更包裹状态 |
| POST | `/pickup/verify` | 核销取件 |
| GET | `/pickup/logs` | 取件日志 |
| GET/POST/PUT | `/shelves*` | 货架 CRUD |
| GET | `/shelves/occupancy` | 占用热力图 |
| GET | `/stats/dashboard` | 首页看板 |
| GET | `/stats/trend` | 包裹趋势 |
| GET | `/stats/proxy-finance` | 代取财务 |

---

## 7. 数据库

### 7.1 表列表（11 张业务表）

| 表 | 说明 | 核心索引 |
|----|------|----------|
| `stations` | 驿站信息 | uk_name |
| `users` | 用户 | uk_phone, uk_openid |
| `admins` | 管理员 | uk_username |
| `runner_applications` | 跑腿员申请 | idx_user, idx_status |
| `parcels` | 包裹（核心表） | uk_tracking_station, uk_pickup_code_station, 8 个索引 |
| `pickup_logs` | 取件日志 | idx_parcel, idx_operator |
| `proxy_orders` | 代取订单 | uk_parcel, idx_station_status |
| `shelf_layout` | 货架布局 | uk_station_shelf, version 乐观锁 |
| `notifications` | 通知记录 | idx_user_read |
| `courier_companies` | 快递公司字典 | uk_company_name |
| `operation_logs` | 操作日志 | idx_module_action |

### 7.2 迁移文件

```
migrations/
├── 20260705100000_init_schema.sql     # 11 张表 DDL
└── 20260705100001_seed_courier_companies.sql  # 8 家快递公司种子数据
```

---

## 8. 测试

### 8.1 测试分层

| 层级 | 位置 | 工具 | 说明 |
|------|------|------|------|
| 单元测试 | `internal/` 各包 | testify | Mock 服务层依赖，纯 Go |
| 集成测试 | `test/` | testcontainers + testify | 真实 MySQL 8 + Redis 7 容器 |
| E2E 测试 | `test/e2e_integration_test.go` | 同上 | 跨模块业务流程 |

### 8.2 测试统计

```
单元测试:  7 包 (config/errors/handler/log/middleware/models/service/router)
集成测试: 84 例 (User:14, Parcel:15, Pickup:12, Proxy:18, Shelf:9, Notify:5, Stats:6, E2E:5)
```

### 8.3 运行测试

```bash
# 单元测试（快，无需 Docker）
make test-unit

# 集成测试（需 Docker）
make test-integration

# 单个测试
go test -tags=integration -run TestParcel_01 ./test/ -v
```

---

## 9. 故障排查 (Troubleshooting)

### 9.1 编译/构建

**`go build` 报 "package is not in std"**
> 确认在 `backend/` 目录内执行，不要从仓库根目录构建。

**`go build` 报 undefined**
> 运行 `go mod tidy` 确保依赖完整。

### 9.2 数据库/Migration

**`make migrate-up` 报 "Access denied"**
> 检查 MySQL 容器是否运行、密码是否正确（默认 `1973`）。也可手动执行：
> ```bash
> docker exec -it pickup-mysql mysql -uroot -p1973 -e "CREATE DATABASE IF NOT EXISTS pickup_helper"
> goose -dir migrations mysql "root:1973@tcp(127.0.0.1:3306)/pickup_helper?parseTime=true&loc=Asia%2FShanghai&charset=utf8mb4" up
> ```

**`make migrate-up` 报 "command not found: goose"**
> 安装 goose：`go install github.com/pressly/goose/v3/cmd/goose@latest`

**数据库连接超时**
> 确保 MySQL 时区正确：`SET GLOBAL time_zone = '+08:00';`

### 9.3 运行时

**启动后 401 "未登录或 Token 失效"**
> 检查 `configs/config.dev.yaml` 中 JWT secret 是否与前端一致。

**启动后 CORS 报错**
> 检查 `configs/config.dev.yaml` 中 `cors.allowed_origins` 是否包含前端域名。`*` 通配符仅在不同时需要 Authorization/Credentials 时不生效。

**端口冲突（`bind: address already in use`）**
> 使用 `APP_PORT=18080 make run` 指定其他端口。检查 8080 是否被前端 agent 占用。

**"货架编号不存在或已满" 但货架表有数据**
> 货架容量已达 `max_capacity`。使用 `GET /shelves/occupancy` 查看占用率，或通过 `PUT /shelves/:id` 扩容。

### 9.4 测试

**集成测试超时（"test timed out after 5m"）**
> 默认超时 300s。84 个集成测试每个约 12s，全部运行需 ~17 分钟。建议按模块分批：
> ```bash
> go test -tags=integration -run 'TestParcel' ./test/... -v
> go test -tags=integration -run 'TestE2E' ./test/... -v
> ```

**集成测试报 "Cannot connect to Docker daemon"**
> 启动 Docker 服务，确保当前用户有权限访问 `/var/run/docker.sock`。

**单个测试失败且容器未清理**
> 清理残留容器：`docker rm -f $(docker ps -aq --filter "ancestor=mysql:8.0")`

### 9.5 取件码生成

**"取件码生成失败" (code=10205)**
> 极端并发下取件码碰撞超过 10 次重试。目前为密码学随机 6 位（10^6 空间），单驿站在途包裹通常 <1000，概率极低。若复现，可检查 `parcels` 表是否已有大量同站待取包裹。

### 9.6 代取订单

**接单报 10413 "无跑腿员资质"**
> 用户 `user_type=2` 且 `runner_status=2` 才可接单。使用 `GET /user/info` 检查当前身份。

**发布代取报 10403 "已有进行中订单"**
> 同一包裹已有 pending/delivering/confirm 状态的订单。一个包裹同时只能有一个活跃代取订单。

### 9.7 JWT/鉴权

**Refresh Token 报 "已过期" (10122)**
> Refresh token 有效期 7 天。过期后需重新登录。

**Access Token 报 "未登录或 Token 失效" (10002)**
> Access token 有效期 2 小时。前端应自动用 refresh token 换新，或引导用户重新登录。

### 9.8 日志

**日志级别调整**
> 设置环境变量 `LOG__LEVEL=error` 减少输出，或 `LOG__LEVEL=debug` 查看详细日志（含 SQL 参数）。

**trace_id 追踪**
> 每个请求自动注入 12 位 hex trace_id，通过响应头 `X-Trace-Id` 和日志 `trace_id` 字段关联。

### 9.9 性能

**慢查询排查**
> MySQL 开启 slow_query_log：`SET GLOBAL slow_query_log = ON; SET GLOBAL long_query_time = 1;`

**连接池耗尽**
> 检查 `max_open_conns`（默认 50）。高并发场景可适当增大并监控 `Threads_connected`。

**Redis 内存**
> 验证码 Key 带 TTL（300s），限流 Key 带 TTL（60s/600s），内存占用很低。极端情况可监控 `used_memory`。

### 9.10 常见 HTTP 状态码映射

| HTTP | code 范围 | 典型错误 |
|------|-----------|----------|
| 400 | 10001, 10203, 10204, 10404, 10405, 10502 | 参数/格式错误 |
| 401 | 10002, 10111, 10112, 10121, 10122, 10161 | 未认证/Token 失效 |
| 403 | 10003, 10104, 10113, 10143, 10413 | 权限不足/黑名单 |
| 404 | 10004, 10151, 10171, 10221, 10411, 10511 | 资源不存在 |
| 409 | 10005, 10141, 10201, 10232, 10403, 10412, 10501 | 状态冲突 |
| 429 | 10006 | 频率超限 |
| 500 | 10009, 99999 | 服务端内部错误 |

---

## 10. 集成服务说明

### 10.1 微信小程序登录

**前提条件：微信小程序必须完成企业认证（非个人主体）。** 个人主体小程序调用 `jscode2session` 会返回 `-1 系统错误` 或权限拒绝。

- 端点：`POST /api/v1/auth/wechat-login`
- 流程：`wx.login()` code → 后端 code2session 换取 openid → 查数据库 → 老用户签发 JWT / 新用户注册
- 首次登录通过 `<button open-type="getPhoneNumber">` 获取手机号自动注册
- 手机号验证码登录 (`POST /api/v1/auth/login`) 保留，供网页端使用

### 10.2 短信验证码

当前为 Stub 实现（`service/sms_stub.go`），dev/test 环境固定返回 `123456`，生产环境生成随机 6 位数字通过日志打印。

**接入腾讯云 SMS 需提供：**

| 配置项 | 说明 | 示例 |
|--------|------|------|
| `SecretId` | 腾讯云 API 密钥 ID | `AKIDxxxxxxxx` |
| `SecretKey` | 腾讯云 API 密钥 Key | `xxxxxxxx` |
| `SmsSdkAppId` | SMS 应用 ID | `1400xxxxxx` |
| `SignName` | 短信签名（需审核通过） | `驿站助手` |
| `TemplateId` | 短信模板 ID（需审核通过） | `1234567` |

接入后替换 `sms_stub.go` 中的 `SMSProvider` 实现即可，不影响调用方代码。
