// 根据角色动态生成路由
import { defineFakeRoute } from "vite-plugin-fake-server/client";

export default defineFakeRoute([
  {
    url: "/api/v1/auth/login",
    method: "post",
    response: ({ body }) => {
      if (body.username === "admin") {
        return {
          code: 0,
          message: "success",
          data: {
            accessToken: "eyJhbGciOiJIUzUxMiJ9.admin",
            refreshToken: "eyJhbGciOiJIUzUxMiJ9.adminRefresh",
            expires: "2030/12/31 23:59:59",
            avatar: "https://avatars.githubusercontent.com/u/44761321",
            username: "admin",
            nickname: "系统管理员",
            roles: ["admin"],
            permissions: ["*:*:*"]
          }
        };
      } else {
        return {
          code: 10001,
          message: "用户名或密码错误"
        };
      }
    }
  },
  {
    url: "/api/v1/auth/refresh",
    method: "post",
    response: () => ({
      code: 0,
      message: "success",
      data: {
        access_token: "eyJhbGciOiJIUzUxMiJ9.admin",
        expires_in: 7200
      }
    })
  },
  {
    url: "/api/v1/user/info",
    method: "get",
    response: () => ({
      code: 0,
      message: "success",
      data: {
        id: 1,
        phone: "138****0000",
        nickname: "系统管理员",
        avatar: "https://avatars.githubusercontent.com/u/44761321",
        user_type: 1,
        runner_status: 0,
        credit_score: 100,
        is_blacklisted: false,
        station_id: null,
        created_at: "2026-01-01T00:00:00+08:00"
      }
    })
  }
]);
