import { create } from 'zustand';
import { proxyApi } from '@/api/proxy';
import type { ProxyTask, ProxyOrder, TaskQueryParams, PublishParams } from '@/api/types';

interface ProxyState {
  taskList: ProxyTask[];
  taskLoading: boolean;
  taskHasMore: boolean;
  taskPage: number;

  myOrders: ProxyOrder[];
  myOrdersLoading: boolean;
  myOrdersHasMore: boolean;
  myOrdersPage: number;

  fetchTasks: (params?: TaskQueryParams, refresh?: boolean) => Promise<void>;
  fetchMyOrders: (role?: 'publisher' | 'taker', refresh?: boolean) => Promise<void>;
  publishTask: (data: PublishParams) => Promise<number>;
  acceptTask: (id: number) => Promise<ProxyOrder>;
  confirmDelivery: (id: number, accepted: boolean, reason?: string) => Promise<void>;
  cancelOrder: (id: number, reason: string) => Promise<void>;
  loadMoreTasks: () => Promise<void>;
  loadMoreOrders: () => Promise<void>;
}

const PAGE_SIZE = 20;

export const useProxyStore = create<ProxyState>((set, get) => ({
  taskList: [],
  taskLoading: false,
  taskHasMore: true,
  taskPage: 1,

  myOrders: [],
  myOrdersLoading: false,
  myOrdersHasMore: true,
  myOrdersPage: 1,

  fetchTasks: async (params, refresh = false) => {
    const state = get();
    if (state.taskLoading) return;

    set({ taskLoading: true });
    try {
      const page = refresh ? 1 : state.taskPage;
      const result = await proxyApi.getTasks({ ...params, page, page_size: PAGE_SIZE });

      set({
        taskList: refresh ? result.list : [...state.taskList, ...result.list],
        taskPage: page,
        taskHasMore: result.list.length >= PAGE_SIZE,
        taskLoading: false,
      });
    } catch {
      set({ taskLoading: false });
    }
  },

  fetchMyOrders: async (role, refresh = false) => {
    const state = get();
    if (state.myOrdersLoading) return;

    set({ myOrdersLoading: true });
    try {
      const page = refresh ? 1 : state.myOrdersPage;
      const result = await proxyApi.getMyOrders({ role, page, page_size: PAGE_SIZE });

      set({
        myOrders: refresh ? result.list : [...state.myOrders, ...result.list],
        myOrdersPage: page,
        myOrdersHasMore: result.list.length >= PAGE_SIZE,
        myOrdersLoading: false,
      });
    } catch {
      set({ myOrdersLoading: false });
    }
  },

  publishTask: async (data) => {
    const result = await proxyApi.publish(data);
    return result.id;
  },

  acceptTask: async (id) => {
    return proxyApi.accept(id);
  },

  confirmDelivery: async (id, accepted, reason) => {
    await proxyApi.confirmDelivery(id, { accepted, reason });
  },

  cancelOrder: async (id, reason) => {
    await proxyApi.cancelOrder(id, reason);
  },

  loadMoreTasks: async () => {
    const state = get();
    if (!state.taskHasMore || state.taskLoading) return;
    set({ taskPage: state.taskPage + 1 });
    await get().fetchTasks();
  },

  loadMoreOrders: async () => {
    const state = get();
    if (!state.myOrdersHasMore || state.myOrdersLoading) return;
    set({ myOrdersPage: state.myOrdersPage + 1 });
    await get().fetchMyOrders();
  },
}));
