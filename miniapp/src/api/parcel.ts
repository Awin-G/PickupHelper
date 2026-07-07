import { request } from './request';
import type { Parcel, PickupCodeInfo, PaginatedList } from './types';

export const parcelApi = {
  /** 我的包裹列表 */
  getMy: (params: { status?: number; page?: number; page_size?: number }) =>
    request<PaginatedList<Parcel>>({
      url: '/parcels/my',
      params,
    }),

  /** 包裹详情 */
  getDetail: (id: number) =>
    request<Parcel>({
      url: `/parcels/${id}`,
    }),

  /** 获取取件码（含二维码） */
  getPickupCode: (id: number) =>
    request<PickupCodeInfo>({
      url: `/parcels/${id}/pickup-code`,
    }),

  /** 待取件数量 */
  getPendingCount: () =>
    request<{ count: number }>({
      url: '/parcels/pending-count',
    }),

  /** 自助出库 */
  selfCheckout: (data: { pickup_code: string; station_id: number }) =>
    request<{ parcel_id: number; pickup_time: string }>({
      url: '/pickup/self-checkout',
      method: 'POST',
      data,
      withGeo: true,
    }),

  /** 扫驿站码批量出库 */
  scanStation: (data: { station_qr: string; pickup_codes: string[] }) =>
    request<{ success: any[]; failed: any[] }>({
      url: '/pickup/scan-station',
      method: 'POST',
      data,
      withGeo: true,
    }),
};
