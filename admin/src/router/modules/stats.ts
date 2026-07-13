import { $t } from "@/plugins/i18n";
import { stats } from "@/router/enums";

export default {
  path: "/stats",
  redirect: "/stats/overview",
  meta: {
    icon: "ep/pie-chart",
    title: "数据统计",
    rank: stats
  },
  children: [
    {
      path: "/stats/overview",
      name: "StatsOverview",
      component: () => import("@/views/stats/overview/index.vue"),
      meta: {
        title: "数据概览",
        icon: "ep/pie-chart"
      }
    },
    {
      path: "/stats/courier",
      name: "StatsCourier",
      component: () => import("@/views/stats/courier/index.vue"),
      meta: {
        title: "快递公司对账",
        icon: "ep/document"
      }
    },
    {
      path: "/stats/finance",
      name: "StatsFinance",
      component: () => import("@/views/stats/finance/index.vue"),
      meta: {
        title: "代取财务",
        icon: "ep/money"
      }
    }
  ]
} satisfies RouteConfigsTable;
