import { $t } from "@/plugins/i18n";
import { shelf } from "@/router/enums";

export default {
  path: "/shelf",
  meta: {
    icon: "ep/grid",
    title: "货架管理",
    rank: shelf
  },
  children: [
    {
      path: "/shelf/list",
      name: "ShelfList",
      component: () => import("@/views/shelf/index.vue"),
      meta: {
        title: "货架列表",
        icon: "ep/grid"
      }
    }
  ]
} satisfies RouteConfigsTable;
