# CONTEXT-backend.md — 后端开发上下文

> **本文件是后端开发的唯一上下文入口。新会话开始时必读，并必须在每次阶段性进展后主动维护更新 + 提交到 git。**

---

## 1. 项目结构

```
PickupHelper/                      # 仓库根目录（前后端共用）
├── backend/                       # ← 后端 Go 项目（本文件管辖范围）
│   ├── cmd/server/                # 入口 main.go + wire.go（手动 DI）
│   ├── internal/                  # 业务代码（三层架构 handler→service→repository）
│   │   ├── config/                # viper 配置加载
│   │   ├── errors/                # 统一错误码 + AppError + HTTP 映射
│   │   ├── handler/               # HTTP 处理器（Gin）
│   │   ├── log/                   # slog 封装 + trace_id
│   │   ├── middleware/            # JWT / CORS / recovery / ratelimit / trace / validator / admin_only
│   │   ├── models/                # 领域模型 + DTO + 常量 + 工具函数（MaskPhone 等）
│   │   ├── repository/            # 数据访问层（sqlx 原生 SQL）+ DBTX 接口 + WithTx
│   │   ├── router/                # 路由注册 + 中间件链编排
│   │   ├── server/                # http.Server 封装
│   │   └── service/               # 业务逻辑层（含 SMS stub）
│   ├── test/                      # 集成测试（testcontainers MySQL+Redis）+ fixtures
│   ├── configs/                   # config.dev.yaml / config.test.yaml
│   ├── migrations/                # goose SQL 迁移脚本
│   ├── Makefile                   # run/build/test/migrate 等命令（从 backend/ 内执行）
│   ├── go.mod / go.sum            # module name = pickup-helper
│   └── .gitignore                 # （继承根 .gitignore 的 Go 规则）
├── 需求规格说明/                   # 需求文档（前后端共享契约）
├── 详细设计文档/                   # 详细设计 + 数据库 DDL + API 契约（后端实现的黄金标准）
├── 可行性分析报告/                 # 可行性分析
├── .planning/                     # GSD 规划产物（本地，已 gitignore）
├── .gitignore                     # 根级 gitignore
└── CONTEXT-backend.md             # ← 本文件
```

**关键约定**：所有后端命令（`make run` / `go test` / `go build`）必须在 `backend/` 目录内执行。

---

## 2. 技术栈

| 维度 | 选型 |
|------|------|
| 语言 | Go 1.26.4 |
| Web 框架 | Gin v1.12 |
| DB | MySQL 8.0 / MariaDB（sqlx 原生 SQL，不用 ORM） |
| 缓存 | Redis 7（go-redis v9） |
| 迁移 | goose v3 |
| 配置 | viper |
| 日志 | log/slog（结构化） |
| 校验 | go-playground/validator v10（自定义 `phone_cn` tag） |
| 鉴权 | JWT（golang-jwt/jwt v5，access + refresh 双 token） |
| 密码 | bcrypt |
| 测试 | testify + testcontainers（MySQL + Redis 真容器集成测试） |
| 架构 | 单体模块化，三层 handler→service→repository，DBTX 接口统一 *sqlx.DB/*sqlx.Tx |

---

## 3. 当前进度（必须随进展更新此节）

### Roadmap 总览（8 个 Phase）

| Phase | 模块 | 状态 | 提交 |
|-------|------|------|------|
| 1 | Foundation（骨架/配置/日志/迁移/中间件/JWT/测试基建） | ✅ Complete | b818979 |
| 2 | User Module（认证/用户/跑腿员审核/黑名单） | ✅ Complete | df20722 |
| 3 | Parcel Module（包裹入库/查询/批量导入） | ✅ Complete | cb9ebce |
| 4 | Pickup Module（取件核销/出库） | ✅ Complete | 3844b79 |
| 5 | Proxy Module（代取增值服务） | ✅ Complete | 9557259 |
| 6 | Shelf Module（货架布局/容量） | ✅ Complete | 5a3d1d2 |
| 7 | Notification Module（通知/异步管道） | ✅ Complete | b6e5e63 |
| 8 | Stats & Cron Module（统计/定时任务） | ✅ Complete | 25e51bc |

### Phase 2 已交付能力

- `POST /api/v1/auth/send-code` — 发送验证码（SMS stub，测试码 123456）
- `POST /api/v1/auth/login` — 手机号验证码登录/注册，签发 access+refresh JWT
- `POST /api/v1/auth/refresh` — refresh token 换新 access token
- `POST /api/v1/admin/auth/login` — 管理员用户名密码登录
- `GET /api/v1/user/info` / `PUT /api/v1/user/info` — 用户信息查询/更新（手机号脱敏 138****0000）
- `POST /api/v1/user/runner/apply` — 跑腿员资质申请
- `GET /api/v1/admin/user/runner/applications` — 管理员查看申请列表
- `PUT /api/v1/admin/user/runner/applications/:id/audit` — 审核跑腿员申请（事务：applications + users 一致性）
- `PUT /api/v1/admin/users/:id/blacklist` — 黑名单管理（AdminOnly 中间件保护）

### Phase 3 已交付能力

- `POST /api/v1/parcels/scan-in` — 包裹入库（扫码/手动），6位取件码生成（密码学随机）、货架自动分配、乐观锁容量管理
- `GET /api/v1/parcels` — 管理员包裹列表（按驿站筛选、多条件过滤、分页）
- `GET /api/v1/parcels/:id` — 包裹详情（管理员/本人权限隔离）
- `PUT /api/v1/parcels/:id/status` — 包裹状态变更（滞留/退件/异常，状态机校验）
- `GET /api/v1/parcels/my` — 我的包裹列表（仅收件人关联包裹）
- `GET /api/v1/parcels/:id/pickup-code` — 获取取件码（仅本人、仅待取状态）
- `repository.ShelfRepo` — 货架分配基础接口（Phase 6 的货架模块前置）

### Phase 4 已交付能力

- `POST /api/v1/pickups/verify` — 取件核销（取件码验证+状态变更）
- `POST /api/v1/pickups/self-checkout` — 自助取件
- `POST /api/v1/pickups/scan-station` — 扫描台取件
- `GET /api/v1/pickups/logs` — 取件日志查询

### Phase 5 已交付能力

- `POST /api/v1/proxy/orders` — 发布代取订单
- `POST /api/v1/proxy/orders/:id/accept` — 跑腿员接单
- `PUT /api/v1/proxy/orders/:id/delivery` — 配送中状态更新
- `PUT /api/v1/proxy/orders/:id/confirm` — 确认完成
- `PUT /api/v1/proxy/orders/:id/cancel` — 取消订单
- `GET /api/v1/proxy/orders/my` — 我的订单列表
- `GET /api/v1/proxy/orders/tasks` — 跑腿员任务列表

### Phase 6 已交付能力

- `POST /api/v1/shelves` — 创建货架
- `GET /api/v1/shelves` — 货架列表
- `PUT /api/v1/shelves/:id` — 更新货架信息
- `DELETE /api/v1/shelves/:id` — 删除货架
- `GET /api/v1/shelves/:id/heatmap` — 货架占用热力图

### Phase 7 已交付能力

- `GET /api/v1/notifications` — 通知列表（分页）
- `PUT /api/v1/notifications/:id/read` — 标记已读

### Phase 8 已交付能力

- `GET /api/v1/stats/dashboard` — 仪表盘统计
- `GET /api/v1/stats/trend` — 趋势统计
- `GET /api/v1/stats/proxy-finance` — 代取财务统计

---

## 4. 如何运行 / 测试

```bash
# 所有命令在 backend/ 内执行
cd backend

# 启动依赖（本地开发，MySQL+Redis 容器）
docker run -d --name pickup-mysql -p 3306:3306 -e MYSQL_ROOT_PASSWORD=1973 -e MYSQL_DATABASE=pickup_helper mysql:8.0
docker run -d --name pickup-redis -p 6379:6379 redis:7-alpine

# 数据库迁移
make migrate-up

# 启动服务（默认 :8080，但开发测试时用 APP_PORT 覆盖，避免与前端冲突）
make run

# 开发/测试时指定非默认端口（推荐 18080），避免前端 agent 也占用 8080
APP_PORT=18080 make run

# 单元测试（不带集成测试）
make test-unit

# 集成测试（需要 Docker，启动 testcontainers MySQL+Redis）
make test-integration

# 静态检查
make vet
make vet-integration
```

**验证一个 Phase 完成的标准**：`go build ./...` + `go vet ./...` + `go vet -tags=integration ./...` + `make test-unit` + `make test-integration` 全部 PASS，且 `make run` 冒烟测试核心接口可用。

### 端口约定（重要）

- **默认端口 `8080` 仅供生产/前端 agent 使用**，后端开发和测试绝不占用。
- 后端开发/测试时通过环境变量 `APP_PORT` 覆盖，推荐 `18080`：
  ```bash
  APP_PORT=18080 make run
  ```
- 集成测试使用 testcontainers（动态端口），不占用宿主机固定端口。
- 该约定写入本文件，后端 agent 每次启动服务前必须检查。

---

## 5. Git 工作流（必读）

### 分支模型

```
master                # 生产分支（由其他 agent 负责从 backend 合入，不要动）
  └── backend         # 后端集成分支（本工作流的稳定主干）
        ├── phase-02-user-module   # 已完成（Phase 2）
        ├── phase-03-parcel        # 新 Phase 从 backend 切出
        └── ...
```

**规则**：
1. **`backend` 分支是后端开发的集成分支**，从当前位置（phase-02-user-module 完成后）创建。
2. **每个新 Phase 从 `backend` 分支切出 feature 分支**，命名 `phase-XX-<module>`（如 `phase-03-parcel`）。
3. **`backend` → `master` 的合并由其他 agent 负责**，后端开发者不要直接向 master 提交或合并。
4. Phase 内的 plan 级别如需隔离，可在 phase 分支下再切 `phase-XX-YY-<topic>` 子分支，完成后合回 phase 分支。
5. Phase 完成后，将 phase 分支合入 `backend`（fast-forward 或 PR 均可），保持 `backend` 始终是最新稳定状态。

### 合并策略（必须遵守）

- **永远使用 `merge`，禁止 `rebase`**。任何分支合入另一个分支时，必须保留 merge commit（`git merge`，不要 `--squash` 也不要 `--rebase`）。
- 分叉时，用 `git merge`（而非 `git pull --rebase`）合并远程变更，保留 merge commit 以维持历史拓扑完整。
- 目的：merge commit 明确记录了"哪个分支在何时以何种方式合入"，`git log --graph --oneline` 可以一眼看出各分支之间的关系和各版本来源。rebase 会把分支历史压成一条直线，销毁了分支拓扑信息，导致无法追踪问题出自哪个分支。

### master 分支保护规则（必须遵守）

- **`master` 分支上禁止任何直接提交**（包括代码、文档、配置修改等一切变更）。`master` 只允许存在 merge commit，即只能从开发分支（如 `backend`）合入。
- 所有修改（包括本文档的更新）必须在开发分支（`backend` 或 phase 分支）上完成并提交，然后 merge 到 `master`。
- 最终效果：`git log --oneline --graph master` 看到的应全部是 merge commit，开发分支的所有实际提交都在各自分支线上。


### 提交规则

每个 Phase 内分多次提交，遵循以下前缀约定（与已有历史一致）：

| 前缀 | 用途 | 示例 |
|------|------|------|
| `feat(phase-XX-YY):` | 功能实现（model/repo/service/handler/middleware/router） | `feat(phase-02-03): add user-module HTTP layer ...` |
| `test(phase-XX-YY):` | 测试代码（与实现分提交） | `test(phase-02-03): add USER-01~14 integration tests` |
| `verify(phase-XX):` | Phase 级验证提交（build/vet/test 全绿 + 冒烟） | `verify(phase-02-03): build/vet/test PASS + smoke` |
| `fix(phase-XX-YY):` | Bug 修复 | `fix(phase-01-03): enforce JWT on unmatched /api/v1/*` |
| `chore:` | 非功能性变更（重构/依赖/结构） | `chore: move backend code into backend/ subfolder` |
| `docs:` | 文档（含本文件） | `docs: update CONTEXT-backend.md for phase-03 start` |

**要点**：
- `feat` 与 `test` 必须分提交（实现代码与测试代码分离），便于 review。
- 每个 plan 完成后立即提交，不要积压多个 plan 一次提交。
- 提交信息体用中文或英文均可，但前缀和 scope 用英文，保持与历史一致。
- 单次提交不要跨 Phase。

---

## 6. 架构与编码约定

### 三层架构

```
handler (Gin)  →  service (业务逻辑)  →  repository (SQL)
     ↑                  ↑                      ↑
  DTO/请求响应      领域逻辑/事务          DBTX 接口（*sqlx.DB / *sqlx.Tx）
```

- **handler**：只做请求绑定/校验/调用 service/组装响应，不含业务逻辑。用 `middleware.BindAndValidate` + `handler.Success` / `handler.Error`。
- **service**：业务逻辑核心，依赖 repo 接口（便于 mock 单测）。事务用 `repository.WithTx`。
- **repository**：纯数据访问，sqlx 原生 SQL，不泄漏 sqlx 类型给上层（用 DBTX 接口）。

### 统一响应

```json
{"code": 0, "msg": "success", "data": {...}, "trace_id": "xxx"}
```

错误码分段：10xxx 通用，101xx User，102xx Parcel，103xx Pickup，104xx Proxy，105xx Shelf，106xx Notify，107xx Stats。详见 [backend/internal/errors/codes.go](backend/internal/errors/codes.go)。

### 中间件链

```
Recovery → TraceID → Logger → CORS → RateLimit → JWT → AdminOnly → Validator
```

### 关键文件指引

- 错误码定义：`backend/internal/errors/codes.go`
- 路由编排：`backend/internal/router/router.go`
- 依赖注入：`backend/cmd/server/wire.go`
- 迁移脚本：`backend/migrations/`（goose up/down）
- 测试夹具：`backend/test/fixtures.go`
- 集成测试基建：`backend/test/setup.go`（testcontainers）

### 设计文档契约

后端实现的黄金标准是 `详细设计文档/` 下的三份文档：
- `数据库设计文档.md` — 11 张表完整 DDL + CRUD SQL（迁移脚本必须与之对齐）
- `api详细设计.md` — 全部接口请求/响应契约、错误码、校验规则（handler 必须与之对齐）
- `快递代取与驿站管理系统详细设计文档.md` — 架构与模块划分

实现时如发现文档与代码冲突，以文档为准并修正代码；如文档有误，先改文档再改代码，并在提交信息说明。

---

## 7. 本文件维护规则（重要）

1. **新会话开始时**：先读本文件，了解当前 Phase 进度与约定，再开始工作。
2. **每个 Phase 完成后**：必须更新第 3 节「当前进度」表格（状态/提交 hash），并补充「已交付能力」清单。
3. **每个 plan 完成后**：可酌情更新第 3 节的 Roadmap 表格进度。
4. **架构/约定有变化时**：同步更新第 6 节。
5. **每次更新本文件后必须提交 git**，提交信息用 `docs: update CONTEXT-backend.md ...`。
6. **不要让本文件与代码状态脱节**——脱节的上下文文件比没有更糟。

---

## 8. 已知问题 / 待办

- Phase 2 集成测试中 `TestHealth_Live_Integration` 曾因 trace_id 中间件覆盖请求头值而失败，已修复（改用 `X-Trace-Id` 请求头传递）。
- SMS provider 目前是 stub（`service/sms_stub.go`），真实接入留待后续。
- 异步管道（通知/批量导入）目前用 goroutine + buffered channel，MQ 接口抽象留待 Phase 7。

---

*最后更新：2026-07-09（Phase 1-8 全部完成，backend 已 merge 到 master）*
