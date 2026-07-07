# 快递驿站助手 - 小程序开发上下文

## 项目概述

**项目名称**: 快递驿站助手 (PickupHelper)
**技术栈**: Taro 4.2.0 + React 18 + TypeScript + SCSS + Zustand + NutUI React
**目标平台**: 微信小程序 (P0)
**后端地址**: https://blossom.awin-x.top (Go + Gin)
**小程序 AppID**: wxdf3eff7d0ee24324

## 总体目标

构建一个快递驿站管理小程序，支持两种角色：
- **收件人**: 查看包裹、取件码、发布代取任务
- **跑腿员**: 浏览任务大厅、接单、配送、确认送达

核心功能：
1. 包裹管理（入库通知、取件码、自助出库）
2. 代取服务（发布任务、接单、配送、确认）
3. 消息通知（入库、滞留、代取状态）
4. 用户系统（登录、角色切换、信用分）

## 页面路由

| 页面 | 路径 | 分包 | 状态 |
|------|------|------|------|
| 首页 | /pages/index/index | 主包 | ✅ 骨架完成 |
| 个人中心 | /pages/mine/index | 主包 | ✅ 骨架完成 |
| 登录 | /pages/login/index | 主包 | ✅ 骨架完成 |
| 包裹详情 | /subpkg-parcel/pages/parcel-detail/index | subpkg-parcel | ⏳ 待实现 |
| 取件码 | /subpkg-parcel/pages/pickup-code/index | subpkg-parcel | ⏳ 待实现 |
| 自助出库 | /subpkg-parcel/pages/self-checkout/index | subpkg-parcel | ⏳ 待实现 |
| 发布代取 | /subpkg-proxy/pages/proxy-publish/index | subpkg-proxy | ⏳ 待实现 |
| 任务大厅 | /subpkg-proxy/pages/proxy-hall/index | subpkg-proxy | ⏳ 待实现 |
| 代取详情 | /subpkg-proxy/pages/proxy-detail/index | subpkg-proxy | ⏳ 待实现 |
| 我的代取订单 | /subpkg-proxy/pages/proxy-orders/index | subpkg-proxy | ⏳ 待实现 |
| 跑腿员申请 | /subpkg-user/pages/runner-apply/index | subpkg-user | ⏳ 待实现 |
| 消息中心 | /subpkg-user/pages/message-center/index | subpkg-user | ⏳ 待实现 |
| 驿站导航 | /subpkg-user/pages/station-map/index | subpkg-user | ⏳ 待实现 |

## 当前进度

### 已完成 ✅

**基础设施**
- [x] Taro 项目初始化 (4.2.0 + React + TS + Sass)
- [x] monorepo .gitignore 配置
- [x] pnpm workspace 配置
- [x] miniprogram-ci 上传/预览能力

**样式体系**
- [x] SCSS 变量 (_variables.scss)
- [x] Mixin (_mixins.scss)
- [x] 样式重置 (_reset.scss)

**工具函数**
- [x] format.ts - 手机号脱敏、金额格式化、时间格式化
- [x] validator.ts - 表单校验
- [x] storage.ts - 本地存储封装
- [x] constants.ts - 状态常量定义

**API 层**
- [x] types.ts - 通用类型定义
- [x] request.ts - 请求封装 (Token/刷新/重试/Mock)
- [x] auth.ts - 鉴权 API
- [x] parcel.ts - 包裹 API
- [x] proxy.ts - 代取 API
- [x] notification.ts - 通知 API
- [x] mock.ts - Mock 数据 (开发环境自动启用)

**状态管理**
- [x] useUserStore - 用户信息、登录态、角色切换
- [x] useParcelStore - 包裹列表、待取件数
- [x] useProxyStore - 任务大厅、我的订单
- [x] useNotificationStore - 消息通知、未读数

**自定义 Hooks**
- [x] useAuth - 登录态守卫
- [x] useGeoLocation - 地理位置
- [x] usePagination - 分页加载
- [x] usePullRefresh - 下拉刷新

**公共组件**
- [x] StatusBadge - 状态标签
- [x] EmptyState - 空状态占位
- [x] ParcelCard - 包裹卡片

**路由配置**
- [x] app.config.ts - 主包 + 3 分包 + preloadRule + tabBar
- [x] TabBar 图标占位文件

### 待实现 ⏳

**页面开发 (按优先级)**
- [x] 首页完善 - 收件人/跑腿员双模式
- [x] 登录页 - 手机号 + 验证码
- [x] 个人中心 - 用户信息、角色切换、功能入口
- [x] 包裹详情 - 状态、操作按钮
- [x] 取件码展示 - 二维码 + 常亮
- [x] 任务大厅 - 列表、筛选、接单
- [x] 发布代取 - 表单、校验
- [x] 代取详情 - 状态机、操作
- [x] 我的代取订单 - 双视角列表
- [x] 消息中心 - 通知列表
- [x] 驿站导航 - 地图
- [x] 跑腿员申请 - 表单
- [x] 自助出库 - 扫码

**功能完善**
- [x] 真实 API 对接 (替换 Mock) - Pickup 模块已集成
- [ ] 微信登录 (wx.login + 手机号)
- [ ] 订阅消息授权
- [ ] 图片上传 (送达照片)
- [ ] 二维码生成 (取件码)
- [ ] 地图导航

**优化**
- [ ] 骨架屏
- [ ] 图片懒加载
- [ ] 虚拟列表 (长列表)

## 开发命令

```bash
# 安装依赖
cd miniapp && pnpm install

# H5 本地调试 (推荐用于 UI 开发)
npx taro build --type h5
cd dist && python3 -m http.server 8888
# 浏览器访问 http://localhost:8888
# 注意：修改代码后需要重启 HTTP 服务器

# 微信小程序编译
pnpm dev:weapp
pnpm build:weapp

# 预览 (需要 AppID 和密钥)
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

## 注意事项

1. **小程序环境不支持可选链 (`?.`)**，需要使用传统写法
2. **开发环境自动启用 Mock 数据**，无需真实后端
3. **TabBar 图标需要替换**，当前是占位文件
4. **API_BASE 配置**在 `miniapp/.env.development` 和 `.env.production`
5. **密钥文件**在 `~/.config/private.wxdf3eff7d0ee24324.key`

## Git 提交规范

```
feat: 新功能
fix: 修复
docs: 文档
style: 样式
refactor: 重构
test: 测试
chore: 构建/工具
```

## 相关文档

- 详细设计文档: `/详细设计文档/小程序端详细设计文档.md`
- API 设计: `/详细设计文档/api详细设计.md`
- 数据库设计: `/详细设计文档/数据库设计文档.md`
