import { $t } from "@/plugins/i18n";
import { proxy } from "@/router/enums";

export default {
  path: "/proxy",
  meta: {
    icon: "ep/van",
    title: "代取管理",
    rank: proxy
  },
  children: [
    {
      path: "/proxy/orders",
      name: "ProxyOrders",
      component: () => import("@/views/proxy/index.vue"),
      meta: {
        title: "代取订单",
        icon: "ep/van"
      }
    }
  ]
} satisfies RouteConfigsTable;
