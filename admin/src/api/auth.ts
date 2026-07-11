import { http } from "@/utils/http";
import type { ApiResponse } from "./types/parcel";

type LoginData = {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  role: string;
};

type RefreshData = {
  access_token: string;
  expires_in: number;
};

type UserInfoData = {
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

/** 管理员登录 */
export const adminLogin = (data: { username: string; password: string }) => {
  return http.request<ApiResponse<LoginData>>("post", "/admin/auth/login", { data });
};

/** 刷新 Token */
export const refreshTokenApi = (data: { refresh_token: string }) => {
  return http.request<ApiResponse<RefreshData>>("post", "/auth/refresh", { data });
};

/** 获取当前用户信息 */
export const getUserInfo = () => {
  return http.request<ApiResponse<UserInfoData>>("get", "/user/info");
};
