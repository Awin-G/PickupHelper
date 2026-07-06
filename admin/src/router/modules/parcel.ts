import { $t } from "@/plugins/i18n";
import { parcel } from "@/router/enums";

export default {
  path: "/parcel",
  redirect: "/parcel/list",
  meta: {
    icon: "ep/box",
    title: "包裹管理",
    rank: parcel
  },
  children: [
    {
      path: "/parcel/inbound",
      name: "ParcelInbound",
      component: () => import("@/views/parcel/inbound/index.vue"),
      meta: {
        title: "包裹入库",
        icon: "ep/upload"
      }
    },
    {
      path: "/parcel/list",
      name: "ParcelList",
      component: () => import("@/views/parcel/list/index.vue"),
      meta: {
        title: "包裹列表",
        icon: "ep/list"
      }
    },
    {
      path: "/parcel/batch-import",
      name: "ParcelBatchImport",
      component: () => import("@/views/parcel/batch-import/index.vue"),
      meta: {
        title: "批量导入",
        icon: "ep/upload-filled"
      }
    }
  ]
} satisfies RouteConfigsTable;
