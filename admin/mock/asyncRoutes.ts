// 模拟后端动态生成路由（静态路由已在 router/modules/ 中定义）
import { defineFakeRoute } from "vite-plugin-fake-server/client";

export default defineFakeRoute([
  {
    url: "/api/v1/get-async-routes",
    method: "get",
    response: () => ({ code: 0, data: [] })
  }
]);
