import { http } from "@/utils/http";
import type {
  ParcelItem,
  ScanInRequest,
  ScanInResponse,
  BatchInRequest,
  BatchInResponse,
  ParcelListParams,
  PaginatedList
} from "./types/parcel";

/** 扫码/手动入库 */
export const scanIn = (data: ScanInRequest) => {
  return http.request<ScanInResponse>("post", "/parcels/scan-in", { data });
};

/** 批量导入包裹 */
export const batchImport = (data: BatchInRequest) => {
  return http.request<BatchInResponse>("post", "/parcels/batch-in", {
    data,
    headers: { "Content-Type": "multipart/form-data" }
  });
};

/** 包裹列表 */
export const getParcelList = (params?: ParcelListParams) => {
  return http.request<PaginatedList<ParcelItem>>("get", "/parcels", { params });
};

/** 包裹详情 */
export const getParcelDetail = (id: number) => {
  const url = `/parcels/${id}`;
  return http.request<ParcelItem>("get", url);
};

/** 更改包裹状态 */
export const updateParcelStatus = (
  id: number,
  data: { status: number; reason?: string }
) => {
  const url = `/parcels/${id}/status`;
  return http.request<ParcelItem>("put", url, { data });
};
