// 模拟后端动态生成路由
import { defineFakeRoute } from "vite-plugin-fake-server/client";

const adminRoutes = [
  {
    path: "/welcome",
    meta: { icon: "ep/home-filled", title: "工作台", rank: 0 }
  },
  {
    path: "/parcel/inbound",
    meta: { icon: "ep/upload", title: "包裹入库", rank: 1 }
  },
  {
    path: "/parcel/list",
    meta: { icon: "ep/list", title: "包裹列表", rank: 1 }
  },
  {
    path: "/parcel/batch-import",
    meta: { icon: "ep/upload-filled", title: "批量导入", rank: 1 }
  },
  {
    path: "/pickup/verify",
    meta: { icon: "ep/check", title: "核销取件", rank: 2 }
  },
  {
    path: "/shelf/list",
    meta: { icon: "ep/grid", title: "货架列表", rank: 3 }
  },
  {
    path: "/proxy/orders",
    meta: { icon: "ep/van", title: "代取订单", rank: 4 }
  },
  { path: "/user/list", meta: { icon: "ep/user", title: "用户列表", rank: 5 } },
  {
    path: "/user/runner/audit",
    meta: { icon: "ep/checked", title: "跑腿员审核", rank: 5 }
  },
  {
    path: "/stats/overview",
    meta: { icon: "ep/pie-chart", title: "数据概览", rank: 6 }
  },
  {
    path: "/stats/courier",
    meta: { icon: "ep/document", title: "快递公司对账", rank: 6 }
  },
  {
    path: "/stats/finance",
    meta: { icon: "ep/money", title: "代取财务", rank: 6 }
  },
  {
    path: "/station/list",
    meta: { icon: "ep/office-building", title: "驿站列表", rank: 7 }
  }
];

export default defineFakeRoute([
  {
    url: "/api/v1/get-async-routes",
    method: "get",
    response: () => ({ code: 0, data: adminRoutes })
  }
]);
