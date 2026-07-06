import { http } from "@/utils/http";

/** 管理员登录 */
export const adminLogin = (data: { username: string; password: string }) => {
  return http.request<LoginResult>("post", "/auth/login", { data });
};

/** 刷新 Token */
export const refreshTokenApi = (data: { refresh_token: string }) => {
  return http.request<RefreshTokenResult>("post", "/auth/refresh", { data });
};

/** 获取当前用户信息 */
export const getUserInfo = () => {
  return http.request<UserInfoResult>("get", "/user/info");
};

export type LoginResult = {
  code: number;
  message: string;
  data: {
    access_token: string;
    refresh_token: string;
    expires_in: number;
    user: {
      id: number;
      phone: string;
      nickname: string;
      avatar: string;
      user_type: number;
      roles: string[];
      permissions: string[];
      station_id: number;
    };
  };
};

export type RefreshTokenResult = {
  code: number;
  message: string;
  data: {
    access_token: string;
    expires_in: number;
  };
};

export type UserInfoResult = {
  code: number;
  message: string;
  data: {
    id: number;
    phone: string;
    nickname: string;
    avatar: string;
    user_type: number;
    runner_status: number;
    credit_score: number;
    is_blacklisted: boolean;
    station_id: number;
    created_at: string;
  };
};
