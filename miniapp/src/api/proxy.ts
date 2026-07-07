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

  /** 代取订单详情 */
  getDetail: (id: number) =>
    request<ProxyOrder>({
      url: `/proxy/orders/${id}`,
    }),

  /** 我的代取订单 */
  getMyOrders: (params: { role?: 'publisher' | 'taker'; page?: number; page_size?: number }) =>
    request<PaginatedList<ProxyOrder>>({
      url: '/proxy/orders/my',
      params,
    }),

  /** 确认送达 */
  confirmDelivery: (id: number, data: { accepted: boolean; reason?: string }) =>
    request<void>({
      url: `/proxy/orders/${id}/confirm`,
      method: 'POST',
      data,
    }),

  /** 取消订单 */
  cancelOrder: (id: number, reason: string) =>
    request<void>({
      url: `/proxy/orders/${id}/cancel`,
      method: 'POST',
      data: { reason },
    }),
};
