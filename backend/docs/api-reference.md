# API 端点速查

> 完整 API 契约见 `详细设计文档/api详细设计.md`（项目根目录）。

## 认证状态图

```
                 ┌── 公开 ──┐               ┌── JWT ──┐              ┌── AdminOnly ──┐
  /auth/*        ✓          ✗              ✗          ✗             ✗               ✗
  /admin/auth/*   ✓          ✗              ✗          ✗             ✗               ✗
  /health*       ✓          ✗              ✗          ✗             ✗               ✗
  /user/*         ✗          ✓              ✗          ✗             ✗               ✗
  /parcels/my     ✗          ✓              ✗          ✗             ✗               ✗
  /parcels/:id    ✗          ✓              ✗          ✗             ✗               ✗ (权限在 handler 内)
  /parcels/*      ✗          ✗              ✗          ✓             ✗               ✗ (管理员)
  /pickup/verify  ✗          ✗              ✗          ✓             ✗               ✗ (handler 内区分 admin/runner)
  /pickup/logs    ✗          ✗              ✗          ✓             ✗               ✗
  /pickup/self-*  ✗          ✓              ✗          ✗             ✗               ✗
  /proxy/*        ✗          ✓              ✗          ✗             ✗               ✗ (handler 内权限)
  /shelves*       ✗          ✗              ✗          ✓             ✗               ✗
  /stats*         ✗          ✗              ✗          ✓             ✗               ✗
  /notifications* ✗          ✓              ✗          ✗             ✗               ✗
```

## 包裹状态流转

```
待取(1) ──verify──→ 已取(2)
   │               
   ├───→ 滞留(3) ──→ 已退件(4)
   │               
   └───→ 异常(5)
```

## 代取订单状态流转

```
待接单(1) ──accept──→ 配送中(2) ──request-delivery──→ 待确认(3) ──confirm──→ 已完成(4)
   │                      │                              │
   ├──cancel──→ 已取消(5)  ├──cancel──→ 取件失败(6)      └──reject──→ 取件失败(6)
```

## 目录结构速查

```
internal/
├── handler/    # 12 个 handler
│   ├── auth_handler.go          → POST /auth/send-code, /auth/login, /auth/refresh
│   │                             → POST /admin/auth/login
│   ├── user_handler.go          → GET/PUT /user/info, POST /user/runner/apply
│   │                             → GET/PUT /admin/user/runner/...
│   ├── parcel_handler.go        → POST /parcels/scan-in, GET /parcels, ...
│   ├── pickup_handler.go        → POST /pickup/verify, /pickup/self-checkout, ...
│   ├── proxy_handler.go         → POST /proxy/publish, /proxy/accept/:id, ...
│   ├── shelf_handler.go         → GET/POST/PUT /shelves, GET /shelves/occupancy
│   ├── notify_handler.go        → GET /notifications, PUT /notifications/read
│   ├── stats_handler.go         → GET /stats/dashboard, /stats/trend, /stats/proxy-finance
│   ├── health.go                 → GET /health, GET /health/ready
│   ├── response.go              → Success/Error/SuccessPaged 辅助函数
│   └── (test files)
├── service/    # 10 个 service
│   ├── auth_service.go
│   ├── user_service.go
│   ├── parcel_service.go
│   ├── pickup_service.go
│   ├── proxy_service.go
│   ├── shelf_service.go
│   ├── notify_service.go
│   ├── stats_service.go
│   ├── sms_stub.go
│   └── (test files)
├── repository/ # 7 个 repo 接口 + tx.go
│   ├── user_repo.go             → UserRepo, AdminRepo, RunnerAppRepo
│   ├── parcel_repo.go           → ParcelRepo, ShelfRepo
│   ├── pickup_repo.go           → PickupLogRepo
│   ├── proxy_repo.go            → ProxyOrderRepo
│   ├── notify_repo.go           → NotifyRepo
│   ├── cache_repo.go            → SMSCodeCache (Redis)
│   └── tx.go                    → DBTX 接口 + WithTx
├── middleware/ # 7 个中间件
│   ├── jwt.go                   → JWTAuth / JWTAuthOptional / CurrentUser
│   ├── admin_only.go            → AdminOnly
│   ├── recovery.go
│   ├── trace.go
│   ├── logger.go
│   ├── cors.go
│   ├── ratelimit.go
│   └── validator.go             → BindAndValidate + phone_cn 自定义校验
├── models/     # 7 个领域模型
│   ├── user.go                  → User, Admin, RunnerApplication + DTO
│   ├── parcel.go                → Parcel + DTO
│   ├── pickup.go                → PickupLog + DTO
│   ├── proxy.go                 → ProxyOrder + DTO
│   ├── notify.go                → Notification + DTO
│   └── (test files)
├── errors/     # 错误码体系
│   ├── codes.go                 → 67 个错误码 + HTTPStatus() + Msg()
│   └── app_error.go             → AppError 类型
├── config/     # 配置
├── log/        # 结构化日志
├── server/     # HTTP Server 封装
└── router/     # 路由编排
```
