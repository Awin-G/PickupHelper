import { create } from 'zustand';
import { storage } from '@/utils/storage';
import { authApi } from '@/api/auth';
import type { UserInfo } from '@/api/types';

type UserRole = 'receiver' | 'runner';

interface UserState {
  token: string | null;
  refreshToken: string | null;
  userInfo: UserInfo | null;
  currentRole: UserRole;
  isLoggedIn: boolean;

  login: (phone: string, code: string) => Promise<void>;
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
    storage.set('token', result.token);
    storage.set('refresh_token', result.refresh_token);
    set({
      token: result.token,
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
      storage.set('token', result.token);
      storage.set('refresh_token', result.refresh_token);
      set({ token: result.token, refreshToken: result.refresh_token });
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
    await authApi.updateProfile(data);
    set((state) => ({
      userInfo: state.userInfo ? { ...state.userInfo, ...data } : null,
    }));
  },
}));
