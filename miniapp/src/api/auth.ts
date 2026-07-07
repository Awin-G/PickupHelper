import { request } from './request';
import Taro from '@tarojs/taro';
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

  /** 上传头像 */
  uploadAvatar: async (filePath: string) => {
    const token = storage.get<string>('token');
    const res = await Taro.uploadFile({
      url: `${process.env.TARO_APP_API_BASE || 'http://localhost:18080/api/v1'}/user/avatar`,
      filePath,
      name: 'file',
      header: {
        Authorization: token ? `Bearer ${token}` : '',
      },
    });
    const data = JSON.parse(res.data);
    if (data.code !== 0) {
      throw new Error(data.msg);
    }
    return data.data as { avatar_url: string };
  },
};

// 需要引入 storage
import { storage } from '@/utils/storage';
