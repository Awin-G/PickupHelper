import { http } from "@/utils/http";
import type { ProxyOrderItem, PaginatedList } from "./types/parcel";

/** 代取订单列表 */
export const getProxyOrders = (params?: {
  role?: string;
  status?: number;
  page?: number;
  page_size?: number;
}) => {
  return http.request<PaginatedList<ProxyOrderItem>>("get", "/proxy/my-orders", {
    params
  });
};
