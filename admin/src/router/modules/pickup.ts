import { $t } from "@/plugins/i18n";
import { pickup } from "@/router/enums";

export default {
  path: "/pickup",
  redirect: "/pickup/verify",
  meta: {
    icon: "ep/check",
    title: "出库核销",
    rank: pickup
  },
  children: [
    {
      path: "/pickup/verify",
      name: "PickupVerify",
      component: () => import("@/views/pickup/index.vue"),
      meta: {
        title: "核销取件",
        icon: "ep/check",
        showParent: true
      }
    }
  ]
} satisfies RouteConfigsTable;
