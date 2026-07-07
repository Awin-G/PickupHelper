import { create } from 'zustand';
import Taro from '@tarojs/taro';
import { storage } from '@/utils/storage';
import { authApi } from '@/api/auth';
import { generateNickname } from '@/utils/format';
import type { UserInfo } from '@/api/types';

type UserRole = 'receiver' | 'runner';

interface UserState {
  token: string | null;
  refreshToken: string | null;
  userInfo: UserInfo | null;
  currentRole: UserRole;
  isLoggedIn: boolean;

  login: (phone: string, code: string) => Promise<void>;
  wechatLogin: (phoneCode?: string) => Promise<void>;
  logout: () => void;
  refreshAuth: () => Promise<boolean>;
  fetchUserInfo: () => Promise<void>;
  switchRole: (role: UserRole) => void;
  updateProfile: (data: Partial<UserInfo>) => Promise<void>;
}

export const useUserStore = create<UserState>((set, get) => ({
  token: storage.get<string>('token'),
  refreshToken: storage.get<string>('refresh_token'),
  userInfo: null,
  currentRole: storage.get<UserRole>('currentRole') || 'receiver',
  isLoggedIn: !!storage.get<string>('token'),

  login: async (phone, code) => {
    const result = await authApi.login({ phone, code });
    storage.set('token', result.access_token);
    storage.set('refresh_token', result.refresh_token);

    let user = result.user;
    if (!user.nickname) {
      const nickname = generateNickname();
      try {
        user = await authApi.updateProfile({ nickname });
      } catch {
        user = { ...user, nickname };
      }
    }

    set({
      token: result.access_token,
      refreshToken: result.refresh_token,
      userInfo: user,
      isLoggedIn: true,
    });
  },

  wechatLogin: async (phoneCode?: string) => {
    // 调用 wx.login 获取 code
    let loginRes;
    try {
      loginRes = await Taro.login();
    } catch (e) {
      console.error('wx.login 失败:', e);
      throw new Error('微信登录接口调用失败');
    }

    if (!loginRes.code) {
      throw new Error('获取微信登录凭证失败');
    }

    const nickname = generateNickname();
    let result;
    try {
      result = await authApi.wechatLogin({
        code: loginRes.code,
        phone_code: phoneCode,
        nickname,
      });
    } catch (e: any) {
      console.error('wechatLogin API 失败:', e);
      throw e;
    }

    storage.set('token', result.access_token);
    storage.set('refresh_token', result.refresh_token);

    set({
      token: result.access_token,
      refreshToken: result.refresh_token,
      userInfo: result.user,
      isLoggedIn: true,
    });
  },

  logout: () => {
    storage.remove('token');
    storage.remove('refresh_token');
    storage.remove('currentRole');
    set({
      token: null,
      refreshToken: null,
      userInfo: null,
      isLoggedIn: false,
      currentRole: 'receiver',
    });
  },

  refreshAuth: async () => {
    const refreshToken = get().refreshToken;
    if (!refreshToken) return false;
    try {
      const result = await authApi.refreshToken(refreshToken);
      storage.set('token', result.access_token);
      set({ token: result.access_token });
      return true;
    } catch {
      get().logout();
      return false;
    }
  },

  fetchUserInfo: async () => {
    try {
      const userInfo = await authApi.getUserInfo();
      set({ userInfo });
    } catch {
      // 获取失败不做处理
    }
  },

  switchRole: (role) => {
    storage.set('currentRole', role);
    set({ currentRole: role });
  },

  updateProfile: async (data) => {
    const user = await authApi.updateProfile(data);
    set({ userInfo: user });
  },
}));
