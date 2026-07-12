import Taro from '@tarojs/taro';
import { storage } from '@/utils/storage';
import { API_ERROR_CODE } from '@/utils/constants';
import type { ApiResponse } from './types';
import { BusinessError } from './types';

const BASE_URL_FALLBACK = 'https://pickup.awin-x.top/api/v1';
let BASE_URL = BASE_URL_FALLBACK;
let USE_MOCK = false;

// 读取环境变量
try {
  if (typeof process !== 'undefined' && process.env && process.env.TARO_APP_API_BASE) {
    BASE_URL = process.env.TARO_APP_API_BASE;
  }
} catch (e) {
  // 小程序环境
}

interface RequestConfig {
  url: string;
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE';
  data?: any;
  params?: Record<string, any>;
  headers?: Record<string, string>;
  showLoading?: boolean;
  withGeo?: boolean;
  retry?: boolean;
}

function matchMockRoute(config: RequestConfig): ((config: RequestConfig) => Promise<any>) | null {
  return null;
}

function getAppVersion(): string {
  try {
    const accountInfo = Taro.getAccountInfoSync();
    return accountInfo.miniProgram.version || '1.0.0';
  } catch {
    return '1.0.0';
  }
}

function buildUrl(url: string, params?: Record<string, any>): string {
  if (!params) return url;
  const query = Object.entries(params)
    .filter(([, v]) => v !== undefined && v !== null)
    .map(([k, v]) => `${encodeURIComponent(k)}=${encodeURIComponent(String(v))}`)
    .join('&');
  return query ? `${url}?${query}` : url;
}

let isRefreshing = false;
let refreshQueue: Array<() => void> = [];

function redirectToLogin() {
  storage.remove('token');
  storage.remove('refresh_token');
  storage.remove('currentRole');
  Taro.reLaunch({ url: '/pages/login/index' });
}

async function tryRefreshToken(): Promise<boolean> {
  const refreshToken = storage.get<string>('refresh_token');
  if (!refreshToken) {
    redirectToLogin();
    return false;
  }

  if (isRefreshing) {
    return new Promise((resolve) => {
      refreshQueue.push(() => resolve(true));
    });
  }

  isRefreshing = true;
  try {
    const res = await Taro.request({
      url: `${BASE_URL}/auth/refresh`,
      method: 'POST',
      data: { refresh_token: refreshToken },
      header: { 'Content-Type': 'application/json' },
    });
    const body = res.data as ApiResponse<{ token: string; refresh_token: string }>;
    if (body.code === 0) {
      storage.set('token', body.data.token);
      storage.set('refresh_token', body.data.refresh_token);
      refreshQueue.forEach((cb) => cb());
      refreshQueue = [];
      return true;
    }
    redirectToLogin();
    return false;
  } catch {
    redirectToLogin();
    return false;
  } finally {
    isRefreshing = false;
  }
}

export async function request<T>(config: RequestConfig): Promise<T> {
  const token = storage.get<string>('token');
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    'X-Client-Type': 'miniapp',
    'X-Client-Version': getAppVersion(),
    ...config.headers,
  };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  if (config.withGeo) {
    try {
      const loc = await Taro.getLocation({ type: 'gcj02' });
      headers['X-Geo-Lat'] = String(loc.latitude);
      headers['X-Geo-Lng'] = String(loc.longitude);
    } catch {}
  }

  if (config.showLoading) {
    Taro.showLoading({ title: '加载中...', mask: true });
  }

  try {
    const response = await Taro.request({
      url: buildUrl(`${BASE_URL}${config.url}`, config.params),
      method: config.method || 'GET',
      data: config.data,
      header: headers,
    });
    const body = response.data as ApiResponse<T>;
    if (body.code !== 0) {
      if (body.code === API_ERROR_CODE.TOKEN_EXPIRED) {
        const refreshed = await tryRefreshToken();
        if (refreshed) return request<T>(config);
        redirectToLogin();
        throw new BusinessError(body.code, body.msg);
      }
      throw new BusinessError(body.code, body.msg);
    }
    return body.data;
  } catch (err) {
    if (err instanceof BusinessError) throw err;
    throw new BusinessError(-1, '网络请求失败，请稍后重试');
  } finally {
    if (config.showLoading) {
      Taro.hideLoading();
    }
  }
}
