import { create } from 'zustand';
import { parcelApi } from '@/api/parcel';
import type { Parcel, PickupCodeInfo } from '@/api/types';

interface ParcelState {
  myParcels: Parcel[];
  pendingCount: number;
  loading: boolean;
  hasMore: boolean;
  currentPage: number;

  fetchMyParcels: (refresh?: boolean) => Promise<void>;
  fetchPendingCount: () => Promise<void>;
  getParcelDetail: (id: number) => Promise<Parcel>;
  getPickupCode: (id: number) => Promise<PickupCodeInfo>;
  loadMore: () => Promise<void>;
}

const PAGE_SIZE = 20;

export const useParcelStore = create<ParcelState>((set, get) => ({
  myParcels: [],
  pendingCount: 0,
  loading: false,
  hasMore: true,
  currentPage: 1,

  fetchMyParcels: async (refresh = false) => {
    const state = get();
    if (state.loading) return;

    set({ loading: true });
    try {
      const page = refresh ? 1 : state.currentPage;
      const result = await parcelApi.getMy({ page, page_size: PAGE_SIZE });

      set({
        myParcels: refresh ? result.list : [...state.myParcels, ...result.list],
        currentPage: page,
        hasMore: result.list.length >= PAGE_SIZE,
        loading: false,
      });
    } catch {
      set({ loading: false });
    }
  },

  fetchPendingCount: async () => {
    try {
      const result = await parcelApi.getPendingCount();
      set({ pendingCount: result.count });
    } catch {
      // 静默失败
    }
  },

  getParcelDetail: async (id) => {
    return parcelApi.getDetail(id);
  },

  getPickupCode: async (id) => {
    return parcelApi.getPickupCode(id);
  },

  loadMore: async () => {
    const state = get();
    if (!state.hasMore || state.loading) return;
    set({ currentPage: state.currentPage + 1 });
    await get().fetchMyParcels();
  },
}));
