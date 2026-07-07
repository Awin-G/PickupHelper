import Taro from '@tarojs/taro';

const STORAGE_PREFIX = 'pickup_';

export const storage = {
  get<T = any>(key: string): T | null {
    try {
      const value = Taro.getStorageSync(STORAGE_PREFIX + key);
      return value || null;
    } catch {
      return null;
    }
  },

  set(key: string, value: any): void {
    try {
      Taro.setStorageSync(STORAGE_PREFIX + key, value);
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
