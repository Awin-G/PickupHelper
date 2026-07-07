import type { Parcel, ProxyTask, ProxyOrder, Notification, UserInfo, PickupCodeInfo, PaginatedList } from './types';

interface RequestConfig {
  url: string;
  method?: string;
  data?: any;
  params?: Record<string, any>;
  headers?: Record<string, string>;
  showLoading?: boolean;
  withGeo?: boolean;
  retry?: boolean;
}

const mockUser: UserInfo = {
  id: 1,
  phone: '138****1234',
  nickname: '张三',
  avatar: '',
  user_type: 2,
  runner_status: 2,
  credit_score: 100,
  is_blacklisted: false,
  created_at: '2026-07-01T00:00:00+08:00',
};

const mockParcels: Parcel[] = [
  {
    id: 1, station_id: 1, tracking_no: 'SF1234567890', courier_company: '顺丰速运',
    shelf_code: 'A-03', pickup_code: '382916', receiver_phone: '13800001234', receiver_name: '张三',
    status: 1, status_text: '待取', weight: 1.2, is_fragile: false, remarks: '',
    storage_time: '2026-07-05T10:30:00Z', pickup_time: null, return_time: null, notify_count: 0,
  },
  {
    id: 2, station_id: 1, tracking_no: 'ZTO9876543210', courier_company: '中通快递',
    shelf_code: 'B-12', pickup_code: '551208', receiver_phone: '13800001234', receiver_name: '张三',
    status: 1, status_text: '待取', weight: 0.8, is_fragile: true, remarks: '易碎品',
    storage_time: '2026-07-06T14:20:00Z', pickup_time: null, return_time: null, notify_count: 1,
  },
  {
    id: 3, station_id: 2, tracking_no: 'YT20240705001', courier_company: '圆通速递',
    shelf_code: 'C-05', pickup_code: '772341', receiver_phone: '13800001234', receiver_name: '张三',
    status: 3, status_text: '滞留', weight: 2.5, is_fragile: false, remarks: '',
    storage_time: '2026-07-02T09:15:00Z', pickup_time: null, return_time: null, notify_count: 3,
  },
  {
    id: 4, station_id: 1, tracking_no: 'JD0012345678', courier_company: '京东物流',
    shelf_code: 'A-08', pickup_code: '449901', receiver_phone: '13800001234', receiver_name: '张三',
    status: 2, status_text: '已取', weight: 1.0, is_fragile: false, remarks: '',
    storage_time: '2026-07-01T16:00:00Z', pickup_time: '2026-07-03T11:20:00Z', return_time: null, notify_count: 0,
  },
];

const mockPickupCode: PickupCodeInfo = {
  pickup_code: '382916',
  qr_url: 'PICKUP:382916:1',
  expire_at: '2026-07-12T23:59:59Z',
};

const mockTasks: ProxyTask[] = [
  {
    id: 1, parcel_id: 1, station_id: 1, station_name: '南门菜鸟驿站',
    reward_amount: 5.0, deadline: '2026-07-07T18:00:00Z', remark: '轻拿轻放', created_at: '2026-07-07T10:00:00Z',
  },
  {
    id: 2, parcel_id: 2, station_id: 2, station_name: '北门驿站',
    reward_amount: 3.0, deadline: '2026-07-08T12:00:00Z', remark: '', created_at: '2026-07-07T09:30:00Z',
  },
  {
    id: 3, parcel_id: 5, station_id: 1, station_name: '南门菜鸟驿站',
    reward_amount: 8.0, deadline: '2026-07-07T15:00:00Z', remark: '加急，很着急用', created_at: '2026-07-07T11:00:00Z',
  },
];

const mockOrders: ProxyOrder[] = [
  {
    id: 1, parcel_id: 1, station_name: '南门菜鸟驿站', publisher_id: 2, publisher_nickname: '李四',
    taker_id: 1, taker_nickname: '张三', reward_amount: 5.0, status: 2, status_text: '配送中',
    temp_pickup_code: '382916', deadline: '2026-07-07T18:00:00Z', delivery_time: null, created_at: '2026-07-07T10:00:00Z',
  },
  {
    id: 2, parcel_id: 3, station_name: '北门驿站', publisher_id: 1, publisher_nickname: '张三',
    taker_id: 5, taker_nickname: '王**', reward_amount: 3.0, status: 3, status_text: '待确认',
    deadline: '2026-07-08T12:00:00Z', delivery_time: '2026-07-07T16:30:00Z', created_at: '2026-07-07T09:30:00Z',
  },
];

const mockNotifications: Notification[] = [
  { id: 1, title: '入库通知', content: '您有一个新包裹到达南门菜鸟驿站，货架: A-03', type: 1, is_read: false, parcel_id: 1, created_at: '2026-07-07T10:30:00Z' },
  { id: 2, title: '滞留提醒', content: '您有包裹已超过3天未取，请尽快取件', type: 2, is_read: false, parcel_id: 3, created_at: '2026-07-07T09:00:00Z' },
  { id: 3, title: '代取接单通知', content: '您的代取任务已被接单，跑腿员: 张**', type: 3, is_read: true, parcel_id: 2, created_at: '2026-07-06T15:20:00Z' },
  { id: 4, title: '系统通知', content: '欢迎使用快递驿站助手！', type: 4, is_read: true, parcel_id: null, created_at: '2026-07-01T00:00:00Z' },
];

function paginate<T>(list: T[], page: number, pageSize: number): PaginatedList<T> {
  const start = (page - 1) * pageSize;
  return { list: list.slice(start, start + pageSize), total: list.length, page, page_size: pageSize };
}

function getParams(config: RequestConfig) {
  return config.params || {};
}

export const mockRoutes: Record<string, (config: RequestConfig) => Promise<any>> = {
  'POST /auth/send-code': async () => ({ expire_in: 300 }),
  'POST /auth/login': async () => ({
    access_token: 'mock_access_token_xxx',
    refresh_token: 'mock_refresh_token_xxx',
    expires_in: 7200,
    user: mockUser,
    role: 'user',
  }),
  'POST /auth/refresh': async () => ({
    access_token: 'mock_access_token_refreshed',
    expires_in: 7200,
  }),
  'GET /user/info': async () => mockUser,
  'PUT /user/info': async () => mockUser,
  'GET /parcels/my': async (config) => {
    const p = getParams(config);
    return paginate(mockParcels, p.page || 1, p.page_size || 20);
  },
  'GET /parcels/pending-count': async () => ({ count: mockParcels.filter((p) => p.status === 1 || p.status === 3).length }),
  'GET /parcels/:id': async (config) => {
    const match = config.url.match(/\/parcels\/(\d+)/);
    const id = match ? Number(match[1]) : 1;
    return mockParcels.find((p) => p.id === id) || mockParcels[0];
  },
  'GET /parcels/:id/pickup-code': async () => mockPickupCode,
  'GET /proxy/tasks': async (config) => {
    const p = getParams(config);
    return paginate(mockTasks, p.page || 1, p.page_size || 20);
  },
  'POST /proxy/publish': async () => ({ id: 99 }),
  'POST /proxy/accept/:id': async () => mockOrders[0],
  'GET /proxy/orders/:id': async () => mockOrders[0],
  'GET /proxy/orders/my': async (config) => {
    const p = getParams(config);
    return paginate(mockOrders, p.page || 1, p.page_size || 20);
  },
  'POST /proxy/orders/:id/confirm': async () => undefined,
  'POST /proxy/orders/:id/cancel': async () => undefined,
  'GET /notifications': async (config) => {
    const p = getParams(config);
    return paginate(mockNotifications, p.page || 1, p.page_size || 20);
  },
  'GET /notifications/unread-count': async () => ({ count: mockNotifications.filter((n) => !n.is_read).length }),
  'POST /notifications/:id/read': async () => undefined,
  'POST /notifications/read-all': async () => undefined,
};
