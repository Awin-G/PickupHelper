import { request } from './request';
import type { LoginParams, LoginResult, UserInfo } from './types';

export const authApi = {
  /** 发送验证码 */
  sendCode: (phone: string) =>
    request<{ expire_in: number }>({
      url: '/auth/send-code',
      method: 'POST',
      data: { phone },
    }),

  /** 登录 */
  login: (params: LoginParams) =>
    request<LoginResult>({
      url: '/auth/login',
      method: 'POST',
      data: params,
    }),

  /** 刷新 Token */
  refreshToken: (refreshToken: string) =>
    request<{ access_token: string; expires_in: number }>({
      url: '/auth/refresh',
      method: 'POST',
      data: { refresh_token: refreshToken },
    }),

  /** 获取用户信息 */
  getUserInfo: () =>
    request<UserInfo>({
      url: '/user/info',
    }),

  /** 更新用户信息 */
  updateProfile: (data: Partial<UserInfo>) =>
    request<UserInfo>({
      url: '/user/info',
      method: 'PUT',
      data,
    }),
};
