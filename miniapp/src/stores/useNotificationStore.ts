import { create } from 'zustand';
import { notificationApi } from '@/api/notification';
import type { Notification } from '@/api/types';

interface NotificationState {
  unreadCount: number;
  notifications: Notification[];
  loading: boolean;
  hasMore: boolean;
  currentPage: number;

  fetchUnreadCount: () => Promise<void>;
  fetchNotifications: (refresh?: boolean) => Promise<void>;
  markAsRead: (id: number) => Promise<void>;
  markAllAsRead: () => Promise<void>;
  loadMore: () => Promise<void>;
}

const PAGE_SIZE = 20;

export const useNotificationStore = create<NotificationState>((set, get) => ({
  unreadCount: 0,
  notifications: [],
  loading: false,
  hasMore: true,
  currentPage: 1,

  fetchUnreadCount: async () => {
    try {
      const result = await notificationApi.getUnreadCount();
      set({ unreadCount: result.count });
    } catch {
      // 静默失败
    }
  },

  fetchNotifications: async (refresh = false) => {
    const state = get();
    if (state.loading) return;

    set({ loading: true });
    try {
      const page = refresh ? 1 : state.currentPage;
      const result = await notificationApi.getList({ page, page_size: PAGE_SIZE });

      set({
        notifications: refresh ? result.list : [...state.notifications, ...result.list],
        currentPage: page,
        hasMore: result.list.length >= PAGE_SIZE,
        loading: false,
      });
    } catch {
      set({ loading: false });
    }
  },

  markAsRead: async (id) => {
    await notificationApi.markAsRead(id);
    set((state) => ({
      notifications: state.notifications.map((n) =>
        n.id === id ? { ...n, is_read: true } : n
      ),
      unreadCount: Math.max(0, state.unreadCount - 1),
    }));
  },

  markAllAsRead: async () => {
    await notificationApi.markAllAsRead();
    set((state) => ({
      notifications: state.notifications.map((n) => ({ ...n, is_read: true })),
      unreadCount: 0,
    }));
  },

  loadMore: async () => {
    const state = get();
    if (!state.hasMore || state.loading) return;
    set({ currentPage: state.currentPage + 1 });
    await get().fetchNotifications();
  },
}));
