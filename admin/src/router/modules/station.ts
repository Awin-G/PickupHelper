import { $t } from "@/plugins/i18n";
import { station } from "@/router/enums";

export default {
  path: "/station",
  redirect: "/station/list",
  meta: {
    icon: "ep/office-building",
    title: "驿站管理",
    rank: station
  },
  children: [
    {
      path: "/station/list",
      name: "StationList",
      component: () => import("@/views/station/index.vue"),
      meta: {
        title: "驿站列表",
        icon: "ep/office-building",
        showParent: true
      }
    }
  ]
} satisfies RouteConfigsTable;
