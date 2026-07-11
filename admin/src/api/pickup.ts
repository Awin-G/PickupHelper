import { http } from "@/utils/http";
import type {
  PickupVerifyRequest,
  PickupVerifyResponse,
  PickupLogItem,
  ApiResponse,
  PagedResponse
} from "./types/parcel";

/** 核销取件 */
export const verifyPickup = (data: PickupVerifyRequest) => {
  return http.request<ApiResponse<PickupVerifyResponse>>("post", "/pickup/verify", { data });
};

/** 取件日志查询 */
export const getPickupLogs = (params?: {
  parcel_id?: number;
  operator_id?: number;
  operator_type?: number;
  start?: string;
  end?: string;
  page?: number;
  page_size?: number;
}) => {
  return http.request<PagedResponse<PickupLogItem>>("get", "/pickup/logs", {
    params
  });
};
