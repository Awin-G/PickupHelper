import type { Parcel, ProxyTask, ProxyOrder, Notification, UserInfo, PickupCodeInfo, PaginatedList } from './types';

/** 模拟延迟 */
const delay = (ms = 300) => new Promise((r) => setTimeout(r, ms));

/** 模拟用户信息 */
export const mockUser: UserInfo = {
  id: 1,
  phone: '13800001234',
  nickname: '张三',
  avatar: '',
  user_type: 2,
  runner_status: 2,
  credit_score: 100,
  is_blacklisted: false,
};

/** 模拟包裹列表 */
export const mockParcels: Parcel[] = [
  {
    id: 1,
    station_id: 1,
    tracking_no: 'SF1234567890',
    courier_company: '顺丰速运',
    shelf_code: 'A-03',
    pickup_code: '382916',
    receiver_phone: '13800001234',
    receiver_name: '张三',
    status: 1,
    status_text: '待取',
    weight: 1.2,
    is_fragile: false,
    remarks: '',
    storage_time: '2026-07-05T10:30:00Z',
    pickup_time: null,
    return_time: null,
    notify_count: 0,
  },
  {
    id: 2,
    station_id: 1,
    tracking_no: 'ZTO9876543210',
    courier_company: '中通快递',
    shelf_code: 'B-12',
    pickup_code: '551208',
    receiver_phone: '13800001234',
    receiver_name: '张三',
    status: 1,
    status_text: '待取',
    weight: 0.8,
    is_fragile: true,
    remarks: '易碎品',
    storage_time: '2026-07-06T14:20:00Z',
    pickup_time: null,
    return_time: null,
    notify_count: 1,
  },
  {
    id: 3,
    station_id: 2,
    tracking_no: 'YT20240705001',
    courier_company: '圆通速递',
    shelf_code: 'C-05',
    pickup_code: '772341',
    receiver_phone: '13800001234',
    receiver_name: '张三',
    status: 3,
    status_text: '滞留',
    weight: 2.5,
    is_fragile: false,
    remarks: '',
    storage_time: '2026-07-02T09:15:00Z',
    pickup_time: null,
    return_time: null,
    notify_count: 3,
  },
  {
    id: 4,
    station_id: 1,
    tracking_no: 'JD0012345678',
    courier_company: '京东物流',
    shelf_code: 'A-08',
    pickup_code: '449901',
    receiver_phone: '13800001234',
    receiver_name: '张三',
    status: 2,
    status_text: '已取',
    weight: 1.0,
    is_fragile: false,
    remarks: '',
    storage_time: '2026-07-01T16:00:00Z',
    pickup_time: '2026-07-03T11:20:00Z',
    return_time: null,
    notify_count: 0,
  },
];

/** 模拟取件码 */
export const mockPickupCode: PickupCodeInfo = {
  pickup_code: '382916',
  qr_url: 'PICKUP:382916:1',
  expire_at: '2026-07-12T23:59:59Z',
};

/** 模拟代取任务 */
export const mockTasks: ProxyTask[] = [
  {
    id: 1,
    parcel_id: 1,
    station_id: 1,
    station_name: '南门菜鸟驿站',
    reward_amount: 5.0,
    deadline: '2026-07-07T18:00:00Z',
    remark: '轻拿轻放',
    created_at: '2026-07-07T10:00:00Z',
  },
  {
    id: 2,
    parcel_id: 2,
    station_id: 2,
    station_name: '北门驿站',
    reward_amount: 3.0,
    deadline: '2026-07-08T12:00:00Z',
    remark: '',
    created_at: '2026-07-07T09:30:00Z',
  },
  {
    id: 3,
    parcel_id: 5,
    station_id: 1,
    station_name: '南门菜鸟驿站',
    reward_amount: 8.0,
    deadline: '2026-07-07T15:00:00Z',
    remark: '加急，很着急用',
    created_at: '2026-07-07T11:00:00Z',
  },
];

/** 模拟代取订单 */
export const mockOrders: ProxyOrder[] = [
  {
    id: 1,
    parcel_id: 1,
    station_name: '南门菜鸟驿站',
    publisher_id: 2,
    publisher_nickname: '李四',
    taker_id: 1,
    taker_nickname: '张三',
    reward_amount: 5.0,
    status: 2,
    status_text: '配送中',
    temp_pickup_code: '382916',
    deadline: '2026-07-07T18:00:00Z',
    delivery_time: null,
    created_at: '2026-07-07T10:00:00Z',
  },
  {
    id: 2,
    parcel_id: 3,
    station_name: '北门驿站',
    publisher_id: 1,
    publisher_nickname: '张三',
    taker_id: 5,
    taker_nickname: '王**',
    reward_amount: 3.0,
    status: 3,
    status_text: '待确认',
    deadline: '2026-07-08T12:00:00Z',
    delivery_time: '2026-07-07T16:30:00Z',
    created_at: '2026-07-07T09:30:00Z',
  },
];

/** 模拟通知 */
export const mockNotifications: Notification[] = [
  {
    id: 1,
    title: '入库通知',
    content: '您有一个新包裹到达南门菜鸟驿站，货架: A-03',
    type: 1,
    is_read: false,
    parcel_id: 1,
    created_at: '2026-07-07T10:30:00Z',
  },
  {
    id: 2,
    title: '滞留提醒',
    content: '您有包裹已超过3天未取，请尽快取件',
    type: 2,
    is_read: false,
    parcel_id: 3,
    created_at: '2026-07-07T09:00:00Z',
  },
  {
    id: 3,
    title: '代取接单通知',
    content: '您的代取任务已被接单，跑腿员: 张**',
    type: 3,
    is_read: true,
    parcel_id: 2,
    created_at: '2026-07-06T15:20:00Z',
  },
  {
    id: 4,
    title: '系统通知',
    content: '欢迎使用快递驿站助手！',
    type: 4,
    is_read: true,
    parcel_id: null,
    created_at: '2026-07-01T00:00:00Z',
  },
];

/** 分页工具 */
export function paginate<T>(list: T[], page: number, pageSize: number): PaginatedList<T> {
  const start = (page - 1) * pageSize;
  return {
    list: list.slice(start, start + pageSize),
    total: list.length,
    page,
    page_size: pageSize,
  };
}
