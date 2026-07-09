# CONTEXT-admin.md — 管理端前端开发上下文

> **本文件是管理端前端开发的唯一上下文入口。新会话开始时必读，并必须在每次阶段性进展后主动维护更新 + 提交到 git。**

---

## 1. 项目结构

```
PickupHelper/                          # 仓库根目录（前后端共用）
├── admin/                             # ← 管理端前端项目（本文件管辖范围）
│   ├── src/                           # 源代码
│   ├── build/                         # Vite 构建配置
│   ├── mock/                          # Mock 数据
│   ├── index.html                     # 入口 HTML
│   ├── package.json                   # 依赖与脚本
│   ├── vite.config.ts                 # Vite 配置
│   └── ...
├── backend/                           # 后端 Go 项目
├── miniapp/                           # 小程序前端
├── 需求规格说明/                       # 需求文档
├── 详细设计文档/                       # 详细设计 + API 契约
└── CONTEXT-admin.md                   # ← 本文件
```

---

## 2. 技术栈

| 维度 | 选型 |
|------|------|
| 框架 | Vue 3 (Composition API) |
| 构建 | Vite 7 |
| 语言 | TypeScript |
| UI 库 | Element Plus |
| CSS | Tailwind CSS + SCSS |
| 状态管理 | Pinia |
| 路由 | Vue Router |
| 请求 | Axios |
| 包管理 | pnpm |
| 代码规范 | ESLint + Prettier + Stylelint + Husky |
| 后端 API | Go + Gin (backend/ 项目) |

---

## 3. 当前进度

### 已完成 ✅

- [x] 项目初始化（vue-pure-admin 脚手架，基于 PureAdmin v7）
- [x] 详细设计文档
- [x] **管理员登录页** — 用户名+密码登录，记住密码，表单校验
- [x] **仪表盘/数据概览** — 看板卡片（今日入库/出库/待取/滞留/异常/代取）+ 动画数字
- [x] **用户管理**
  - 用户列表（筛选/搜索/分页/黑名单管理）
  - 跑腿员审核（申请列表/证件预览/批准/拒绝）
- [x] **包裹管理**
  - 包裹入库（扫码模式/手动模式，表单校验）
  - 包裹列表（多条件筛选/分页/状态标签/详情弹窗）
  - 批量导入（Excel 上传）
- [x] **取件核销** — 扫码/手动核销 + 最近核销记录
- [x] **货架管理** — CRUD + 占用率可视化（热力图）
- [x] **代取订单管理** — 订单列表/状态筛选/分页
- [x] **统计报表**
  - 数据概览（ECharts 趋势图）
  - 快递公司对账
  - 代取财务汇总
- [x] **驿站管理** — 驿站列表/新增/编辑（系统管理员）
- [x] 自定义组件：`ParcelStatusTag` / `ShelfSelector` / `CourierCompanySelect`
- [x] API 接口层：auth / parcel / pickup / proxy / shelf / stats / station / user（8 个模块）
- [x] 路由模块配置（9 个路由文件，对应完整菜单体系）
- [x] 类型定义：完整的请求/响应 TypeScript 类型
- [x] 工具函数：手机号脱敏 / 日期格式化 / 金额格式化 / 校验函数

### 待实现 / 增强 ⏳

- [ ] **通知管理页面** — 通知列表/手动发送/模板管理（后端接口已就绪）
- [ ] **管理员账号管理** — 子账号新增/编辑/角色分配（账户设置页面已有骨架）
- [ ] 货架占用热力图可视化增强（ECharts 热力图渲染）
- [ ] 与真实后端 API 联调对接（当前使用 mock/proxy 模式）
- [ ] 包裹详情弹窗扩展（完整信息展示 + 操作日志时间线）
- [ ] 取件日志管理页面（独立查看所有取件操作审计日志）
- [ ] 代取订单详情/操作面板（订单完整状态流转监控）
- [ ] E2E 测试覆盖（Vitest + Playwright）

---

## 4. 如何运行

```bash
cd admin

# 安装依赖
pnpm install

# 开发（默认端口由 Vite 分配，通常为 5173）
pnpm dev

# 构建
pnpm build

# 类型检查
pnpm typecheck

# 代码检查
pnpm lint
```

---

## 5. Git 工作流（必读）

### 分支模型

```
master                                # 生产分支（只含 merge commit，禁止直接提交）
  ├── backend                         # 后端集成分支
  ├── feature/admin-frontend-design   # 管理端前端开发分支（本分支）
  └── feature/miniapp-design          # 小程序开发分支
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

每个功能分多次提交，遵循以下前缀约定：

| 前缀 | 用途 | 示例 |
|------|------|------|
| `feat:` | 功能实现 | `feat: 添加用户管理页面` |
| `fix:` | Bug 修复 | `fix: 修复登录页表单校验` |
| `docs:` | 文档 | `docs: 更新 CONTEXT-admin.md` |
| `style:` | 样式 | `style: 调整仪表盘布局` |
| `refactor:` | 重构 | `refactor: 提取通用表格组件` |
| `test:` | 测试 | `test: 添加登录页单元测试` |
| `chore:` | 构建/工具 | `chore: 升级 Element Plus 版本` |

---

## 6. 架构与编码约定

### 目录结构

```
admin/src/
├── api/                 # API 接口层（8 个模块文件）
│   ├── types/parcel.ts  #   完整的类型定义
│   ├── auth.ts          #   认证
│   ├── parcel.ts        #   包裹
│   ├── pickup.ts        #   取件
│   ├── proxy.ts         #   代取
│   ├── shelf.ts         #   货架
│   ├── station.ts       #   驿站
│   ├── stats.ts         #   统计
│   └── user.ts          #   用户
├── assets/              # 静态资源
├── components/          # 公共组件（29 个目录）
│   ├── ParcelStatusTag/ #   包裹状态标签
│   ├── ShelfSelector/   #   货架选择器
│   ├── CourierCompanySelect/ # 快递公司选择器
│   └── ...              #   PureAdmin 内置组件
├── layout/              # 布局组件（PureAdmin 内置）
├── plugins/             # 插件配置
├── router/              # 路由配置
│   └── modules/         #   9 个模块路由文件
├── store/               # Pinia 状态管理
├── utils/               # 工具函数（http/format/validate）
└── views/               # 页面视图（11 个模块，~7000+ 行代码）
    ├── login/           #   登录页
    ├── welcome/         #   工作台/仪表盘
    ├── parcel/          #   包裹管理（入库/列表/批量导入）
    ├── pickup/          #   出库核销
    ├── shelf/           #   货架管理
    ├── proxy/           #   代取订单
    ├── user/            #   用户管理（列表/跑腿员审核）
    ├── stats/           #   数据统计（概览/快递对账/财务）
    ├── station/         #   驿站管理
    └── account-settings/#   账户设置（骨架）
```

### 实现统计

| 指标 | 数值 |
|------|------|
| Vue SFC 文件 | 14 个视图文件 |
| API 模块 | 9 个（含类型定义） |
| 路由模块 | 9 个 |
| 自定义组件 | 3 个 |
| 总代码量 | ~7000+ 行 TypeScript/Vue |

### 统一响应

后端返回格式：
```json
{"code": 0, "msg": "success", "data": {...}, "trace_id": "xxx"}
```

错误码分段：10xxx 通用，101xx User，102xx Parcel，103xx Pickup，104xx Proxy，105xx Shelf，106xx Notify，107xx Stats。

### 设计文档契约

前端实现的黄金标准是 `详细设计文档/` 下的文件：
- `管理端前端详细设计文档.md` — 页面设计 + 交互流程
- `api详细设计.md` — 接口请求/响应契约
- `数据库设计文档.md` — 数据模型参考

---

## 7. 本文件维护规则

1. **新会话开始时**：先读本文件，了解当前进度与约定，再开始工作。
2. **每个功能完成后**：更新第 3 节进度。
3. **架构/约定有变化时**：同步更新第 6 节。
4. **每次更新本文件后必须提交 git**，提交信息用 `docs: update CONTEXT-admin.md ...`。
5. **不要让本文件与代码状态脱节**——脱节的上下文文件比没有更糟。

---

*最后更新：2026-07-09（初始创建，7.0 全部核心模块实现完成）*
