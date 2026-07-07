import Taro from '@tarojs/taro';
import { storage } from '@/utils/storage';
import { API_ERROR_CODE } from '@/utils/constants';
import type { ApiResponse } from './types';
import { BusinessError } from './types';
import { mockRoutes } from './mock';

const BASE_URL_FALLBACK = 'http://localhost:18080/api/v1';
let BASE_URL = BASE_URL_FALLBACK;
let USE_MOCK = true;

// H5 模式下通过 URL 参数或 localStorage 控制 mock
if (typeof window !== 'undefined') {
  const urlParams = new URLSearchParams(window.location.search);
  const mockParam = urlParams.get('mock');
  if (mockParam === 'false') {
    USE_MOCK = false;
  }
  // 检查 localStorage
  const storedMock = localStorage.getItem('pickup_use_mock');
  if (storedMock === 'false') {
    USE_MOCK = false;
  }
}

try {
  if (typeof process !== 'undefined' && process.env && process.env.TARO_APP_API_BASE) {
    BASE_URL = process.env.TARO_APP_API_BASE;
    USE_MOCK = false;
  }
} catch (e) {
  // H5 环境
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
  const method = config.method || 'GET';
  const key = `${method} ${config.url}`;
  if (mockRoutes[key]) return mockRoutes[key];

  for (const [pattern, handler] of Object.entries(mockRoutes)) {
    const [pMethod, pPath] = pattern.split(' ');
    if (pMethod !== method) continue;
    const regex = new RegExp('^' + pPath.replace(/:(\w+)/g, '(?<$1>\\d+)') + '$');
    if (regex.test(config.url)) return handler;
  }
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

async function tryRefreshToken(): Promise<boolean> {
  const refreshToken = storage.get<string>('refresh_token');
  if (!refreshToken) return false;

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
    return false;
  } catch {
    return false;
  } finally {
    isRefreshing = false;
  }
}

export async function request<T>(config: RequestConfig): Promise<T> {
  if (USE_MOCK) {
    const handler = matchMockRoute(config);
    if (handler) {
      await new Promise((r) => setTimeout(r, 200));
      return handler(config) as T;
    }
    console.warn(`[Mock] No mock for: ${config.method || 'GET'} ${config.url}`);
    return {} as T;
  }

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
