import { request } from './request';
import type { ProxyTask, ProxyOrder, PaginatedList, TaskQueryParams, PublishParams } from './types';

export const proxyApi = {
  /** 任务大厅列表 */
  getTasks: (params?: TaskQueryParams) =>
    request<PaginatedList<ProxyTask>>({
      url: '/proxy/tasks',
      params,
    }),

  /** 发布代取任务 */
  publish: (data: PublishParams) =>
    request<{ id: number }>({
      url: '/proxy/publish',
      method: 'POST',
      data,
    }),

  /** 接单 */
  accept: (id: number) =>
    request<ProxyOrder>({
      url: `/proxy/accept/${id}`,
      method: 'POST',
    }),

  /** 请求配送（跑腿员标记开始配送） */
  requestDelivery: (id: number) =>
    request<void>({
      url: `/proxy/request-delivery/${id}`,
      method: 'POST',
    }),

  /** 确认送达/收货 */
  confirmDelivery: (id: number, data: { accepted: boolean; reason?: string }) =>
    request<void>({
      url: `/proxy/confirm-delivery/${id}`,
      method: 'POST',
      data,
    }),

  /** 取消订单 */
  cancelOrder: (data: { order_id: number; reason: string }) =>
    request<void>({
      url: '/proxy/cancel',
      method: 'POST',
      data,
    }),

  /** 我的代取订单 */
  getMyOrders: (params?: { role?: 'publisher' | 'taker'; page?: number; page_size?: number }) =>
    request<PaginatedList<ProxyOrder>>({
      url: '/proxy/my-orders',
      params,
    }),
};
