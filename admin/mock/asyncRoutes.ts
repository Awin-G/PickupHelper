// 模拟后端动态生成路由
import { defineFakeRoute } from "vite-plugin-fake-server/client";

/**
 * roles：页面级别权限，这里模拟二种 "admin"、"common"
 * admin：管理员角色
 * common：普通角色
 */

const adminRoutes = [
  {
    path: "/parcel/inbound",
    meta: {
      icon: "ep/upload",
      title: "包裹入库",
      rank: 1
    }
  },
  {
    path: "/parcel/list",
    meta: {
      icon: "ep/list",
      title: "包裹列表",
      rank: 1
    }
  }
];

export default defineFakeRoute([
  {
    url: "/getAsyncRoutes",
    method: "get",
    response: () => {
      return {
        code: 0,
        data: adminRoutes
      };
    }
  }
]);
