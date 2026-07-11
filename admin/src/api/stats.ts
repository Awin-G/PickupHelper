import { http } from "@/utils/http";
import type { ApiResponse } from "./types/parcel";

type DashboardData = {
  date: string;
  station_id: number;
  today_inbound: number;
  today_outbound: number;
  pending_count: number;
  delayed_count: number;
  abnormal_count: number;
  proxy_active: number;
  shelf_usage_rate: number;
};

type TrendData = {
  granularity: string;
  points: Array<{
    date: string;
    inbound: number;
    outbound: number;
    delayed: number;
  }>;
};

type FinanceData = {
  total_orders: number;
  completed_orders: number;
  total_reward: number;
  avg_reward: number;
  by_taker: Array<{
    taker_id: number;
    taker_nickname: string;
    order_count: number;
    total_reward: number;
  }>;
};

type CourierCheckItem = {
  courier_company: string;
  inbound_count: number;
  outbound_count: number;
  delayed_count: number;
  returned_count: number;
  avg_storage_hours: number;
};

/** 首页看板数据 */
export const getDashboard = (params?: {
  station_id?: number;
  date?: string;
}) => {
  return http.request<ApiResponse<DashboardData>>("get", "/stats/dashboard", { params });
};

/** 包裹趋势图 */
export const getStatsTrend = (params: {
  station_id?: number;
  granularity: "day" | "week" | "month" | "year";
  start?: string;
  end?: string;
}) => {
  return http.request<ApiResponse<TrendData>>("get", "/stats/trend", { params });
};

/** 代取财务汇总 */
export const getProxyFinance = (params?: {
  station_id?: number;
  start?: string;
  end?: string;
}) => {
  return http.request<ApiResponse<FinanceData>>("get", "/stats/proxy-finance", { params });
};

/** 快递公司对账 */
export const getCourierCheck = (params?: {
  station_id?: number;
  courier_company?: string;
  start?: string;
  end?: string;
}) => {
  return http.request<ApiResponse<CourierCheckItem[]>>("get", "/stats/courier-check", { params });
};
