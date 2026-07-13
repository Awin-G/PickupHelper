import { $t } from "@/plugins/i18n";
import { user } from "@/router/enums";

export default {
  path: "/user",
  redirect: "/user/list",
  meta: {
    icon: "ep/user",
    title: "用户管理",
    rank: user
  },
  children: [
    {
      path: "/user/list",
      name: "UserList",
      component: () => import("@/views/user/list/index.vue"),
      meta: {
        title: "用户列表",
        icon: "ep/user"
      }
    },
    {
      path: "/user/runner/audit",
      name: "RunnerAudit",
      component: () => import("@/views/user/runner/audit/index.vue"),
      meta: {
        title: "跑腿员审核",
        icon: "ep/checked"
      }
    }
  ]
} satisfies RouteConfigsTable;
