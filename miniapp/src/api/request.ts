import Taro from '@tarojs/taro';
import { storage } from '@/utils/storage';
import { API_ERROR_CODE } from '@/utils/constants';
import type { ApiResponse } from './types';
import { BusinessError } from './types';
import * as mock from './mock';

const BASE_URL = process.env.TARO_APP_API_BASE || 'http://localhost:8080/api/v1';

// 是否使用 mock 数据（开发环境默认开启）
const USE_MOCK = process.env.NODE_ENV === 'development';

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

/** Mock 路由表 */
const mockRoutes: Record<string, (config: RequestConfig) => Promise<any>> = {
  'POST /auth/send-code': async () => undefined,
  'POST /auth/login': async () => ({
    token: 'mock_token_xxx',
    refresh_token: 'mock_refresh_xxx',
    user: mock.mockUser,
  }),
  'POST /auth/refresh': async () => ({
    token: 'mock_token_refreshed',
    refresh_token: 'mock_refresh_refreshed',
  }),
  'GET /user/info': async () => mock.mockUser,
  'PUT /user/profile': async () => undefined,
  'GET /parcels/my': async (config) => {
    const params = config.params || {};
    return mock.paginate(mock.mockParcels, params.page || 1, params.page_size || 20);
  },
  'GET /parcels/pending-count': async () => ({
    count: mock.mockParcels.filter((p) => p.status === 1 || p.status === 3).length,
  }),
  'GET /parcels/:id': async (config) => {
    const match = config.url.match(/\/parcels\/(\d+)/);
    const id = match ? Number(match[1]) : 1;
    return mock.mockParcels.find((p) => p.id === id) || mock.mockParcels[0];
  },
  'GET /parcels/:id/pickup-code': async () => mock.mockPickupCode,
  'GET /proxy/tasks': async (config) => {
    const params = config.params || {};
    return mock.paginate(mock.mockTasks, params.page || 1, params.page_size || 20);
  },
  'POST /proxy/publish': async () => ({ id: 99 }),
  'POST /proxy/accept/:id': async () => mock.mockOrders[0],
  'GET /proxy/orders/:id': async () => mock.mockOrders[0],
  'GET /proxy/orders/my': async (config) => {
    const params = config.params || {};
    return mock.paginate(mock.mockOrders, params.page || 1, params.page_size || 20);
  },
  'POST /proxy/orders/:id/confirm': async () => undefined,
  'POST /proxy/orders/:id/cancel': async () => undefined,
  'GET /notifications': async (config) => {
    const params = config.params || {};
    return mock.paginate(mock.mockNotifications, params.page || 1, params.page_size || 20);
  },
  'GET /notifications/unread-count': async () => ({
    count: mock.mockNotifications.filter((n) => !n.is_read).length,
  }),
  'POST /notifications/:id/read': async () => undefined,
  'POST /notifications/read-all': async () => undefined,
};

function matchMockRoute(config: RequestConfig): ((config: RequestConfig) => Promise<any>) | null {
  const method = config.method || 'GET';
  const key = `${method} ${config.url}`;

  // 精确匹配
  if (mockRoutes[key]) return mockRoutes[key];

  // 模式匹配 (含 :id)
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
  // Mock 模式
  if (USE_MOCK) {
    const handler = matchMockRoute(config);
    if (handler) {
      await mock.delay(200);
      return handler(config) as T;
    }
    console.warn(`[Mock] No mock for: ${config.method || 'GET'} ${config.url}`);
    return {} as T;
  }

  // 真实请求
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
    } catch {
      // 用户拒绝授权
    }
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
