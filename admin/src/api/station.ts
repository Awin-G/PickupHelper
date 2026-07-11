import { http } from "@/utils/http";
import type { StationItem, StationFormRequest, ApiResponse, PagedResponse } from "./types/parcel";

/** 驿站列表 */
export const getStationList = (params?: {
  keyword?: string;
  status?: number;
  page?: number;
  page_size?: number;
}) => {
  return http.request<PagedResponse<StationItem>>("get", "/stations", { params });
};

/** 新增驿站 */
export const createStation = (data: StationFormRequest) => {
  return http.request<ApiResponse<StationItem>>("post", "/stations", { data });
};

/** 编辑驿站 */
export const updateStation = (id: number, data: Partial<StationFormRequest>) => {
  const url = `/stations/${id}`;
  return http.request<ApiResponse<StationItem>>("put", url, { data });
};
