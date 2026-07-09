# 快递驿站助手 - 小程序开发上下文

## 项目概述

**项目名称**: 快递驿站助手 (PickupHelper)
**技术栈**: Taro 4.2.0 + React 18 + TypeScript + SCSS + Zustand + NutUI React
**目标平台**: 微信小程序 (P0)
**后端地址**: https://pickup.awin-x.top (Go + Gin)
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
- [x] request.ts - 请求封装 (Token/刷新/重试)
- [x] auth.ts - 鉴权 API (含微信登录、头像上传)
- [x] parcel.ts - 包裹 API
- [x] pickup.ts - 取件核销 API
- [x] proxy.ts - 代取 API
- [x] notification.ts - 通知 API

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
- [x] Skeleton - 骨架屏

**路由配置**
- [x] app.config.ts - 主包 + 3 分包 + preloadRule + tabBar
- [x] TabBar 图标占位文件

**页面开发**
- [x] 首页 - 收件人/跑腿员双模式、骨架屏
- [x] 登录页 - 手机号验证码登录
- [x] 个人中心 - 用户信息、角色切换
- [x] 包裹详情 - 状态、操作按钮
- [x] 取件码展示 - 数字展示、页面常亮
- [x] 自助出库 - 扫码、取件码输入
- [x] 发布代取 - 表单、校验
- [x] 任务大厅 - 排序筛选、接单
- [x] 代取详情 - 状态机、操作
- [x] 我的代取订单 - 双视角列表
- [x] 消息中心 - 通知列表
- [x] 驿站导航 - 驿站列表
- [x] 跑腿员申请 - 表单
- [x] 个人信息编辑 - 头像上传、昵称修改

### 搁置 ⏸

| 功能 | 原因 |
|------|------|
| 微信一键登录 | 个人主体小程序无 `getPhoneNumber` 权限 |

### 待实现 ⏳

- [ ] 订阅消息授权
- [ ] 二维码生成 (取件码页面 QR 码)
- [ ] 地图导航 (驿站导航页)

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
npx taro build --type weapp

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
2. **TabBar 图标需要替换**，当前是占位文件
3. **API_BASE 配置**在 `miniapp/.env.development` 和 `.env.production`
4. **密钥文件**在 `~/.config/private.wxdf3eff7d0ee24324.key`
5. **后端地址**: `https://pickup.awin-x.top`

## Git 工作流（必读）

### 分支模型

```
master                                # 生产分支（只含 merge commit，禁止直接提交）
  ├── backend                         # 后端集成分支
  ├── feature/admin-frontend-design   # 管理端前端开发分支
  └── feature/miniapp-design          # 小程序开发分支（本分支）
```

### 合并策略（必须遵守）

- **永远使用 `merge`，禁止 `rebase`**。任何分支合入另一个分支时，必须保留 merge commit（`git merge`，不要 `--squash` 也不要 `--rebase`）。
- 分叉时，用 `git merge`（而非 `git pull --rebase`）合并远程变更，保留 merge commit 以维持历史拓扑完整。
- 目的：merge commit 明确记录了"哪个分支在何时以何种方式合入"，`git log --graph --oneline` 可以一眼看出各分支之间的关系和各版本来源。rebase 会把分支历史压成一条直线，销毁了分支拓扑信息，导致无法追踪问题出自哪个分支。

### master 分支保护规则（必须遵守）

- **`master` 分支上禁止任何直接提交**（包括代码、文档、配置修改等一切变更）。`master` 只允许存在 merge commit，即只能从开发分支合入。
- 所有修改（包括本文档的更新）必须在本分支上完成并提交，然后 merge 到 `master`。
- 最终效果：`git log --oneline --graph master` 看到的应全部是 merge commit，开发分支的所有实际提交都在各自分支线上。

### 提交规则

| 前缀 | 用途 | 示例 |
|------|------|------|
| `feat:` | 新功能 | `feat: 添加订阅消息授权` |
| `fix:` | 修复 | `fix: 修复扫码枪输入` |
| `docs:` | 文档 | `docs: 更新 CONTEXT-miniapp.md` |
| `style:` | 样式 | `style: 调整包裹卡片间距` |
| `refactor:` | 重构 | `refactor: 提取 ParcelStatus 组件` |
| `test:` | 测试 | `test: 添加登录页测试` |
| `chore:` | 构建/工具 | `chore: 升级 Taro 版本` |
| `debug:` | 调试代码 | `debug: 接单流程日志（上线前移除）` |

## 相关文档

- 前端开发文档: `miniapp/FRONTEND.md`
- 详细设计文档: `/详细设计文档/小程序端详细设计文档.md`
- API 设计: `/详细设计文档/api详细设计.md`
- 数据库设计: `/详细设计文档/数据库设计文档.md`
- 后端文档: `backend/docs/README.md`
