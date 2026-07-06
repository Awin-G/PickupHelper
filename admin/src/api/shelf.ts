import { http } from "@/utils/http";
import type { ShelfItem, ShelfFormRequest, PaginatedList } from "./types/parcel";

/** 货架列表 */
export const getShelfList = (params?: {
  station_id?: number;
  page?: number;
  page_size?: number;
}) => {
  return http.request<PaginatedList<ShelfItem>>("get", "/shelves", { params });
};

/** 新增货架 */
export const createShelf = (data: ShelfFormRequest) => {
  return http.request<ShelfItem>("post", "/shelves", { data });
};

/** 更新货架 */
export const updateShelf = (id: number, data: Partial<ShelfFormRequest>) => {
  const url = `/shelves/${id}`;
  return http.request<ShelfItem>("put", url, { data });
};

/** 货架占用热力图数据 */
export const getShelfOccupancy = (params: { station_id: number }) => {
  return http.request<{
    station_id: number;
    shelves: Array<{
      shelf_code: string;
      row_num: number;
      col_num: number;
      current_capacity: number;
      max_capacity: number;
      occupancy_rate: number;
      heat_level: number;
    }>;
    total_used: number;
    total_max: number;
  }>("get", "/shelves/occupancy", { params });
};
