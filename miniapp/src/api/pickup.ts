import { request } from './request';

/** 核销取件请求 */
export interface VerifyPickupParams {
  pickup_code: string;
  verification_method: number; // 1-扫码, 2-手动输入
  station_id: number;
}

/** 核销取件响应 */
export interface VerifyPickupResult {
  parcel_id: number;
  tracking_no: string;
  pickup_time: string;
  operator_type: number;
  proxy_order_id?: number;
}

/** 自助出库请求 */
export interface SelfCheckoutParams {
  pickup_code: string;
  station_id: number;
}

/** 自助出库响应 */
export interface SelfCheckoutResult {
  parcel_id: number;
  pickup_time: string;
}

/** 扫驿站码出库请求 */
export interface ScanStationParams {
  station_qr: string;
  pickup_codes: string[];
}

/** 扫驿站码出库响应 */
export interface ScanStationResult {
  success: Array<{
    pickup_code: string;
    parcel_id: number;
    pickup_time: string;
  }>;
  failed: Array<{
    pickup_code: string;
    reason_code: number;
    reason_msg: string;
  }>;
}

export const pickupApi = {
  /** 核销取件（管理员/跑腿员） */
  verify: (data: VerifyPickupParams) =>
    request<VerifyPickupResult>({
      url: '/pickup/verify',
      method: 'POST',
      data,
      withGeo: true,
    }),

  /** 用户自助出库 */
  selfCheckout: (data: SelfCheckoutParams) =>
    request<SelfCheckoutResult>({
      url: '/pickup/self-checkout',
      method: 'POST',
      data,
      withGeo: true,
    }),

  /** 扫驿站码批量出库 */
  scanStation: (data: ScanStationParams) =>
    request<ScanStationResult>({
      url: '/pickup/scan-station',
      method: 'POST',
      data,
      withGeo: true,
    }),
};
