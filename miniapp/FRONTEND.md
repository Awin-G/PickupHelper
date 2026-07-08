# 快递驿站助手 - 小程序前端开发文档

## 1. 项目概览

### 1.1 技术栈

| 维度 | 选型 | 版本 |
|------|------|------|
| 框架 | Taro | 4.2.0 |
| 语法 | React 18 (Hooks) | 18.x |
| 语言 | TypeScript | 5.4+ |
| UI 组件库 | NutUI React | 3.x |
| 状态管理 | Zustand | 5.x |
| 样式 | SCSS + BEM | — |
| 构建 | Webpack 5 | Taro 内置 |

### 1.2 目录结构

```
miniapp/
├── src/
│   ├── api/                    # API 请求层
│   │   ├── types.ts            # 通用类型定义
│   │   ├── request.ts          # 请求封装 (Token/刷新/重试)
│   │   ├── auth.ts             # 鉴权 API
│   │   ├── parcel.ts           # 包裹 API
│   │   ├── pickup.ts           # 取件核销 API
│   │   ├── proxy.ts            # 代取 API
│   │   └── notification.ts     # 通知 API
│   ├── stores/                 # Zustand 状态管理
│   │   ├── useUserStore.ts     # 用户信息、登录态
│   │   ├── useParcelStore.ts   # 包裹列表、待取件数
│   │   ├── useProxyStore.ts    # 任务大厅、我的订单
│   │   └── useNotificationStore.ts
│   ├── pages/                  # 主包页面
│   │   ├── index/              # 首页
│   │   ├── login/              # 登录页
│   │   └── mine/               # 个人中心
│   ├── subpkg-parcel/          # 包裹分包
│   │   ├── pages/parcel-detail/
│   │   ├── pages/pickup-code/
│   │   └── pages/self-checkout/
│   ├── subpkg-proxy/           # 代取分包
│   │   ├── pages/proxy-publish/
│   │   ├── pages/proxy-hall/
│   │   ├── pages/proxy-detail/
│   │   └── pages/proxy-orders/
│   ├── subpkg-user/            # 用户分包
│   │   ├── pages/runner-apply/
│   │   ├── pages/message-center/
│   │   ├── pages/station-map/
│   │   └── pages/profile-edit/
│   ├── components/             # 公共组件
│   │   ├── ParcelCard/
│   │   ├── StatusBadge/
│   │   ├── EmptyState/
│   │   └── Skeleton/
│   ├── hooks/                  # 自定义 Hooks
│   │   ├── useAuth.ts
│   │   ├── useGeoLocation.ts
│   │   ├── usePagination.ts
│   │   └── usePullRefresh.ts
│   ├── utils/                  # 工具函数
│   │   ├── format.ts
│   │   ├── validator.ts
│   │   ├── storage.ts
│   │   └── constants.ts
│   ├── styles/                 # 全局样式
│   │   ├── _variables.scss
│   │   ├── _mixins.scss
│   │   └── _reset.scss
│   ├── assets/                 # 静态资源
│   │   └── icons/
│   ├── app.ts                  # 应用入口
│   ├── app.config.ts           # 路由配置
│   └── app.scss                # 全局样式
├── config/                     # Taro 编译配置
├── .env.development            # 开发环境变量
├── .env.production             # 生产环境变量
└── package.json
```

---

## 2. 开发规范

### 2.1 Git 提交规范

```
<type>: <description>

类型：
  feat:     新功能
  fix:      修复 Bug
  docs:     文档
  style:    样式调整（不影响逻辑）
  refactor: 重构
  test:     测试
  chore:    构建/工具/依赖
  debug:    调试代码（上线前移除）
```

**示例：**
```
feat: 实现包裹详情页
fix: 修复待取件计数不更新
docs: 更新 CONTEXT.md
style: 优化首页卡片间距
refactor: 移除 mock 模块，改用真实 API
chore: 配置 API 地址为公网域名
debug: 添加微信登录请求日志
```

### 2.2 代码规范

**禁止使用可选链 (`?.`)**
```typescript
// ❌ 错误
if (userInfo?.avatar) { }
const name = user?.nickname || '';

// ✅ 正确
if (userInfo && userInfo.avatar) { }
const name = user ? user.nickname : '';
```

**原因：** 微信小程序基础库不完全支持可选链语法，编译后会报错。

**组件命名：**
- 文件夹名：`kebab-case`（如 `parcel-card`）
- 组件文件：`index.tsx`、`index.scss`
- 导出名：`PascalCase`（如 `ParcelCard`）

**样式命名：**
```scss
// BEM 命名
.parcel-card {
  &__header { }
  &__body { }
  &__footer { }
  &__btn {
    &--primary { }
    &--default { }
  }
}
```

### 2.3 状态管理规范

- 使用 Zustand，按业务域拆分 Store
- Store 放在 `src/stores/` 目录
- 命名：`useXxxStore.ts`
- 每个 Store 包含状态 + 操作方法

```typescript
export const useParcelStore = create<ParcelState>((set, get) => ({
  myParcels: [],
  loading: false,
  fetchMyParcels: async (refresh) => { ... },
}));
```

### 2.4 API 层规范

- 放在 `src/api/` 目录
- 每个模块一个文件（`auth.ts`、`parcel.ts` 等）
- 统一使用 `request` 函数封装
- 类型定义放在 `types.ts`

```typescript
export const parcelApi = {
  getMy: (params) => request<PaginatedList<Parcel>>({ url: '/parcels/my', params }),
  getDetail: (id) => request<Parcel>({ url: `/parcels/${id}` }),
};
```

---

## 3. 环境配置

### 3.1 API 地址

| 环境 | 文件 | API_BASE |
|------|------|----------|
| 开发 | `.env.development` | `https://pickup.awin-x.top/api/v1` |
| 生产 | `.env.production` | `https://pickup.awin-x.top/api/v1` |

### 3.2 小程序配置

| 配置 | 值 |
|------|-----|
| AppID | `wxdf3eff7d0ee24324` |
| 上传密钥 | `~/.config/private.wxdf3eff7d0ee24324.key` |

---

## 4. 开发命令

```bash
# 安装依赖
cd miniapp && pnpm install

# H5 本地调试
npx taro build --type h5
cd dist && python3 -m http.server 8888
# 浏览器访问 http://localhost:8888
# 注意：修改代码后需要重启 HTTP 服务器

# 微信小程序编译
npx taro build --type weapp

# 预览
npx miniprogram-ci preview \
  --appid wxdf3eff7d0ee24324 \
  --project-path . \
  --pkp ~/.config/private.wxdf3eff7d0ee24324.key \
  --uv 1.0.0 \
  --qrcode-format image \
  --qrcode-output-dest /tmp/miniapp-preview.jpg

# 上传
npx miniprogram-ci upload \
  --appid wxdf3eff7d0ee24324 \
  --project-path . \
  --pkp ~/.config/private.wxdf3eff7d0ee24324.key \
  --uv 1.0.0 \
  --desc "版本说明"
```

---

## 5. 页面路由

| 页面 | 路径 | 分包 | 状态 |
|------|------|------|------|
| 首页 | `/pages/index/index` | 主包 | ✅ |
| 登录 | `/pages/login/index` | 主包 | ✅ |
| 个人中心 | `/pages/mine/index` | 主包 | ✅ |
| 包裹详情 | `/subpkg-parcel/pages/parcel-detail/index` | subpkg-parcel | ✅ |
| 取件码 | `/subpkg-parcel/pages/pickup-code/index` | subpkg-parcel | ✅ |
| 自助出库 | `/subpkg-parcel/pages/self-checkout/index` | subpkg-parcel | ✅ |
| 发布代取 | `/subpkg-proxy/pages/proxy-publish/index` | subpkg-proxy | ✅ |
| 任务大厅 | `/subpkg-proxy/pages/proxy-hall/index` | subpkg-proxy | ✅ |
| 代取详情 | `/subpkg-proxy/pages/proxy-detail/index` | subpkg-proxy | ✅ |
| 我的代取订单 | `/subpkg-proxy/pages/proxy-orders/index` | subpkg-proxy | ✅ |
| 跑腿员申请 | `/subpkg-user/pages/runner-apply/index` | subpkg-user | ✅ |
| 消息中心 | `/subpkg-user/pages/message-center/index` | subpkg-user | ✅ |
| 驿站导航 | `/subpkg-user/pages/station-map/index` | subpkg-user | ✅ |
| 个人信息编辑 | `/subpkg-user/pages/profile-edit/index` | subpkg-user | ✅ |

---

## 6. API 对接状态

### 6.1 已对接端点

| 模块 | 端点 | 说明 |
|------|------|------|
| Auth | `POST /auth/send-code` | 发送验证码 |
| Auth | `POST /auth/login` | 手机号登录 |
| Auth | `POST /auth/wechat-login` | 微信登录 |
| Auth | `POST /auth/refresh` | 刷新 Token |
| User | `GET /user/info` | 获取用户信息 |
| User | `PUT /user/info` | 更新用户信息 |
| User | `POST /user/avatar` | 上传头像 |
| User | `GET /user/avatar` | 获取头像 |
| Parcel | `GET /parcels/my` | 我的包裹 |
| Parcel | `GET /parcels/:id` | 包裹详情 |
| Parcel | `GET /parcels/:id/pickup-code` | 取件码 |
| Pickup | `POST /pickup/self-checkout` | 自助出库 |
| Pickup | `POST /pickup/scan-station` | 扫码出库 |
| Proxy | `GET /proxy/tasks` | 任务大厅 |
| Proxy | `POST /proxy/publish` | 发布代取 |
| Proxy | `POST /proxy/accept/:id` | 接单 |
| Proxy | `POST /proxy/confirm-delivery/:id` | 确认收货 |
| Proxy | `GET /proxy/my-orders` | 我的订单 |
| Notify | `GET /notifications` | 通知列表 |
| Notify | `PUT /notifications/read` | 标记已读 |

### 6.2 未实现功能

| 功能 | 原因 | 状态 |
|------|------|------|
| 微信一键登录 | 个人主体小程序无 `getPhoneNumber` 权限 | ⏸ 搁置 |
| 订阅消息授权 | 需要模板 ID | ⏳ 待实现 |
| 二维码生成 | 取件码页面 QR 码 | ⏳ 待实现 |
| 地图导航 | 驿站导航页 | ⏳ 待实现 |

---

## 7. 后端信息

| 项目 | 值 |
|------|-----|
| 后端地址 | `https://pickup.awin-x.top` |
| API 前缀 | `/api/v1` |
| 后端仓库 | `backend/` 分支 |
| 后端技术栈 | Go + Gin + MySQL + Redis |
| 后端端点数 | 34 个 |
| 后端测试数 | 84 个集成测试 |

---

## 8. 已知问题

1. **H5 模式 TabBar 图标不显示** — Taro H5 模式限制，小程序正常
2. **可选链语法** — 禁止使用 `?.`，小程序不支持
3. **mock 模块已移除** — 现在直接连接真实 API
4. **待取件数** — 从包裹列表计算，无独立 API

---

## 9. 相关文档

- 详细设计文档: `/详细设计文档/小程序端详细设计文档.md`
- API 设计: `/详细设计文档/api详细设计.md`
- 数据库设计: `/详细设计文档/数据库设计文档.md`
- 后端文档: `backend/docs/README.md`
