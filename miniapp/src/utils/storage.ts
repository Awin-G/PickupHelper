import Taro from '@tarojs/taro';

const STORAGE_PREFIX = 'pickup_';

export const storage = {
  get<T = any>(key: string): T | null {
    try {
      // H5 模式下直接用 localStorage 绕过
      if (typeof window !== 'undefined' && window.localStorage) {
        const raw = window.localStorage.getItem(STORAGE_PREFIX + key);
        if (!raw) return null;
        try {
          return JSON.parse(raw) as T;
        } catch {
          return raw as T;
        }
      }
      const value = Taro.getStorageSync(STORAGE_PREFIX + key);
      if (!value) return null;
      // Taro H5 模式下会把字符串包装成 {data: value}
      if (typeof value === 'object' && value.data !== undefined) {
        return value.data as T;
      }
      return value as T;
    } catch {
      return null;
    }
  },

  set(key: string, value: any): void {
    try {
      // H5 模式下 Taro 会包装成 {data: value}，直接用 localStorage 绕过
      if (typeof window !== 'undefined' && window.localStorage) {
        window.localStorage.setItem(STORAGE_PREFIX + key, JSON.stringify(value));
      } else {
        Taro.setStorageSync(STORAGE_PREFIX + key, value);
      }
    } catch {
      console.error(`Storage set failed: ${key}`);
    }
  },

  remove(key: string): void {
    try {
      Taro.removeStorageSync(STORAGE_PREFIX + key);
    } catch {
      console.error(`Storage remove failed: ${key}`);
    }
  },

  clear(): void {
    try {
      Taro.clearStorageSync();
    } catch {
      console.error('Storage clear failed');
    }
  },
};
