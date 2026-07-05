# 快递代取与驿站管理系统（后端）

## What This Is

面向校园/社区快递驿站的全流程管理后端服务，以"包裹"为数据核心，以"入库→通知→出库"为主轴，代取服务作为增值插件。基于 Go (Gin + sqlx) 单体模块化架构，为 Taro 小程序（收件人/跑腿员）与 PureAdmin 管理后台提供 RESTful API。

本 GSD 工作流的交付物为**后端代码实现**：设计文档（可行性/需求/详细设计/数据库/API）已完整存在于项目根目录的 `可行性分析报告/`、`需求规格说明/`、`详细设计文档/` 三个子目录，作为后端实现的契约，不再重复设计。

## Core Value

包裹"入库→通知→出库"主流程稳定可靠运行，覆盖 95% 自取场景；代取增值服务作为补充，整体保障驿站日常运营效率与用户取件体验。

## Requirements

### Validated

（无 — 代码尚未实现，待 ship 后验证）

### Active

后端 8 大模块全部实现，每个功能必须配套测试代码：

- [ ] **基础设施**：项目骨架、配置加载（viper）、结构化日志（slog）、错误处理中间件、JWT 鉴权中间件、链路 trace_id、统一响应结构、数据库连接池、Redis 客户端、goose 迁移脚本
- [ ] **用户模块（USER）**：手机号验证码登录/注册（验证码 stub）、JWT 签发与刷新、用户信息查询/更新、跑腿员资质申请、管理员审核、黑名单管理
- [ ] **包裹模块（PARCEL）**：扫码/手动入库、批量 Excel 导入、取件码生成（6位，驿站内唯一）、货架分配、包裹列表（管理员/收件人视角）、包裹详情、状态变更、我的包裹、取件码获取
- [ ] **取件核销模块（PICKUP）**：核销取件（扫码/手动）、用户自助出库（确认码+地理位置校验）、扫驿站二维码出库、取件日志查询、防错拿双重校验、批量取件风控
- [ ] **代取模块（PROXY）**：发布代取任务（悬赏）、任务大厅（排序筛选）、接单（临时授权码生成）、配送确认（双方确认）、收益记录、订单状态流转、超时取消
- [ ] **货架模块（SHELF）**：货架布局 CRUD、实时容量更新（乐观锁）、滞留件自动迁移、占用可视化
- [ ] **通知模块（NOTIFY）**：入库通知、滞留催取（频次控制）、代取状态通知、管理员告警；微信订阅消息/SMS 全部走接口 stub；异步通过 goroutine + buffered channel，预留 MQ 接口
- [ ] **统计模块（STATS）**：包裹流量看板（日/周/月/年）、代取财务统计、快递公司对账、操作日志审计
- [ ] **定时任务模块（CRON）**：超时检测与状态更新（24h 催取/72h 滞留/7d 退件）、代取订单超时取消、每日数据备份、报表生成；robfig/cron + Redis 分布式锁
- [ ] **测试覆盖**：每个功能模块配套单元测试（service 层 mock）+ 集成测试（testcontainers MySQL+Redis），测试代码与实现分提交（`test:` / `verify:`）

### Out of Scope

- **Taro 小程序前端** — 本工作流仅后端，前端不在范围
- **PureAdmin 管理后台前端** — 同上
- **IoT 自助扫码终端** — 设计文档明确"预留 API 不做自助扫码机器"
- **真实外部服务接入** — 微信订阅消息、阿里云 SMS、地图服务全部走接口 stub，真实接入留待后续工作流
- **真实 RabbitMQ/Kafka 部署** — v1 用 goroutine + buffered channel 实现异步，仅抽象 MQ 接口
- **读写分离/分库分表** — 单实例 MySQL 即可，留待后续大促优化
- **微服务拆分** — 单体模块化，后续可拆
- **OAuth 第三方登录** — 设计文档明确 v1 仅手机号验证码
- **CI/CD 流水线** — 不在本工作流范围，本地 dev + 手动测试为主
- **生产部署/容器化** — 不在本工作流范围

## Context

### 已有设计文档（契约）

项目根目录已包含完整的工程设计文档，作为后端实现的契约：

| 文档 | 路径 | 行数 | 内容 |
|------|------|------|------|
| 可行性分析报告 | `可行性分析报告/可行性分析报告.md` | 452 | 业务/技术/经济/社会可行性 |
| 需求规格说明书 | `需求规格说明/快递代取与驿站管理系统需求规格说明书.md` | 200 | 用户角色、功能需求、业务流、数据模型 |
| 详细设计文档 | `详细设计文档/快递代取与驿站管理系统详细设计文档.md` | 980 | 架构、模块划分、ER 图、表结构、接口总览 |
| 数据库设计文档 | `详细设计文档/数据库设计文档.md` | 1307 | 11 张表完整 DDL + 全部 CRUD SQL + 事务/并发/索引策略 |
| API 详细设计 | `详细设计文档/api详细设计.md` | 1376 | 全部接口请求/响应契约、错误码、校验规则 |

### 业务流主线

1. **包裹入库与自取（95% 场景）**：快递员卸货 → 管理员扫码/手动入库 → 系统分配取件码+货架号 → 微信通知收件人 → 收件人到店出示取件码 → 管理员扫码核销/自助扫码 → 出库完成
2. **代取场景（<5%）**：用户不便自取 → 发布代取悬赏 → 跑腿员大厅抢单 → 获临时授权码 → 到店代取 → 送达确认 → 赏金结算

### 用户角色

- **驿站管理员**：包裹入库上架、货架管理、出库核销、异常件处理、营收对账
- **普通收件人**：接收通知、查询取件码、自取、发起代取
- **跑腿员（兼职）**：浏览代取任务、接单配送、收益查看
- **系统管理员**：系统配置、用户权限、全局监控

### 数据模型

11 张表：`stations` / `users` / `admins` / `runner_applications` / `parcels`（核心）/ `pickup_logs` / `proxy_orders` / `shelf_layout`（乐观锁）/ `notifications` / `courier_companies` / `operation_logs`。完整 DDL 与 SQL 见 `详细设计文档/数据库设计文档.md`。

### 本地开发环境

- 本机已有 MariaDB（root/1973），可暂代 MySQL 用于本地 dev 与冒烟测试
- testcontainers 集成测试会自动启 MySQL/Redis 容器，不污染本机数据
- 不要动项目目录以外的任何东西

## Constraints

- **Tech stack**: Go 1.20+ / Gin / sqlx（原生 SQL，无 ORM）/ MySQL 8.0 / Redis 7.0
- **Architecture**: 简洁三层（handler → service → repository），`cmd/` + `internal/` 目录结构
- **Logging**: 标准库 `log/slog`（Go 1.21+）
- **Config**: `viper`（yaml + env 覆盖）
- **Migration**: `goose`（SQL 文件，可逆迁移）
- **Validation**: `go-playground/validator`
- **JWT**: `golang-jwt/jwt/v5`
- **Cron**: `robfig/cron/v3` + Redis 分布式锁
- **Testing**: `testcontainers-go`（MySQL + Redis 真实容器）+ `testify` + service 层 mock
- **Async**: goroutine + buffered channel，MQ 接口抽象（不引真实 MQ）
- **External services**: 微信订阅消息 / SMS / 地图全部走接口 stub
- **Test commit policy**: 每个功能必须配套测试；测试代码与实现同 PR、分提交（`feat:` / `test:` 或 `verify:`）
- **Design contract**: 严格遵循已有设计文档（DDL、SQL、API 契约、错误码），如有冲突以设计文档为准并记录到 Key Decisions

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| 简洁三层架构（handler/service/repository） | 上手快、契合 sqlx 原生 SQL 风格、避免 DDD 模板代码 | — Pending |
| 全部 8 模块完整实现 | 用户明确要求覆盖所有模块 | — Pending |
| 异步用 goroutine + buffered channel，抽象 MQ 接口 | v1 不引真实 MQ，降低部署复杂度，保留升级路径 | — Pending |
| 外部服务全 stub + 接口抽象 | 便于本地开发与测试，后续接真实服务只换实现 | — Pending |
| testcontainers MySQL + Redis 跑集成测试 | 最接近生产，避免 sqlmock 拦截脆弱、避免 sqlite 语法差异 | — Pending |
| slog + viper + goose + validator | 标准库 slog + 生态成熟组件 | — Pending |
| 实现与测试同 PR、分提交（feat: / test:） | git log 可追溯测试单独进化，同时保持 PR 完整性 | — Pending |
| Standard 粒度（5-8 phases） | 平衡 phase 切换成本与单 phase 大小 | — Pending |
| 严格遵循设计文档契约 | DDL/SQL/API/错误码已完整，避免实现层随意偏离 | — Pending |
| 本机 MariaDB 仅用于 dev/冒烟，集成测试走 testcontainers | 不污染本机数据，保证测试可重现 | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-07-05 after initialization*
