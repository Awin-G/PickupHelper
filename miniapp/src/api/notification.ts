import { request } from './request';
import type { Notification, PaginatedList } from './types';

export const notificationApi = {
  /** 通知列表 */
  getList: (params?: { page?: number; page_size?: number }) =>
    request<PaginatedList<Notification>>({
      url: '/notifications',
      params,
    }),

  /** 未读数量 */
  getUnreadCount: () =>
    request<{ count: number }>({
      url: '/notifications/unread-count',
    }),

  /** 标记已读 */
  markAsRead: (id: number) =>
    request<void>({
      url: `/notifications/${id}/read`,
      method: 'POST',
    }),

  /** 全部已读 */
  markAllAsRead: () =>
    request<void>({
      url: '/notifications/read-all',
      method: 'POST',
    }),
};
