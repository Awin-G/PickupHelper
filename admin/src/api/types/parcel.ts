/** API统一响应包装 */
export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

/** 通用分页数据 */
export interface PaginatedList<T> {
  list: T[];
  total: number;
  page: number;
  page_size: number;
}

/** 分页API响应 */
export type PagedResponse<T> = ApiResponse<PaginatedList<T>>;

/** 包裹信息 */
export interface ParcelItem {
  id: number;
  station_id?: number;
  tracking_no: string;
  courier_company: string;
  shelf_code: string;
  pickup_code?: string;
  receiver_phone: string;
  receiver_name?: string;
  weight?: number;
  is_fragile?: boolean;
  remarks?: string;
  status: number;
  status_text?: string;
  storage_time: string;
  pickup_time?: string;
  return_time?: string;
  notify_count: number;
}

/** 扫码入库请求 */
export interface ScanInRequest {
  tracking_no: string;
  courier_company: string;
  receiver_phone: string;
  receiver_name?: string;
  shelf_code?: string;
  weight?: number;
  is_fragile?: boolean;
  remarks?: string;
}

/** 扫码入库响应 */
export interface ScanInResponse {
  parcel_id: number;
  pickup_code: string;
  shelf_code: string;
  storage_time: string;
}

/** 批量导入请求 */
export interface BatchInRequest {
  file: File;
  station_id: number;
}

/** 批量导入响应 */
export interface BatchInResponse {
  batch_id: string;
  total: number;
  status: string;
}

/** 包裹列表查询参数 */
export interface ParcelListParams {
  tracking_no?: string;
  receiver_phone?: string;
  status?: number;
  courier_company?: string;
  shelf_code?: string;
  storage_start?: string;
  storage_end?: string;
  page?: number;
  page_size?: number;
}

/** 核销取件请求 */
export interface PickupVerifyRequest {
  pickup_code: string;
  verification_method: 1 | 2;
  station_id: number;
}

/** 核销取件响应 */
export interface PickupVerifyResponse {
  parcel_id: number;
  tracking_no: string;
  pickup_time: string;
  operator_type: number;
  proxy_order_id?: number;
}

/** 取件日志 */
export interface PickupLogItem {
  id: number;
  parcel_id: number;
  operator_id: number;
  operator_type: number;
  verification_method: number;
  location_lat?: number;
  location_lng?: number;
  ip_address?: string;
  created_at: string;
}

/** 货架信息 */
export interface ShelfItem {
  id: number;
  station_id: number;
  shelf_code: string;
  row_num: number;
  col_num: number;
  current_capacity: number;
  max_capacity: number;
  occupancy_rate: number;
  remark?: string;
}

/** 货架表单请求 */
export interface ShelfFormRequest {
  station_id: number;
  shelf_code: string;
  row_num: number;
  col_num: number;
  max_capacity: number;
  remark?: string;
}

/** 代取订单 */
export interface ProxyOrderItem {
  id: number;
  parcel_id: number;
  station_name?: string;
  publisher_id?: number;
  publisher_nickname?: string;
  taker_id?: number;
  taker_nickname?: string;
  reward_amount: number;
  status: number;
  status_text?: string;
  deadline: string;
  delivery_time?: string;
  created_at: string;
}

/** 用户信息 */
export interface UserItem {
  id: number;
  phone: string;
  nickname: string;
  avatar?: string;
  user_type: number;
  runner_status?: number;
  credit_score?: number;
  is_blacklisted: boolean;
  created_at: string;
}

/** 跑腿员申请 */
export interface RunnerApplicationItem {
  id: number;
  user_id: number;
  real_name: string;
  phone: string;
  student_id?: string;
  id_card_image: string;
  status: number;
  audit_remark?: string;
  created_at: string;
  updated_at: string;
}

/** 跑腿员审核请求 */
export interface RunnerAuditRequest {
  action: "approve" | "reject";
  audit_remark?: string;
}

/** 驿站信息 */
export interface StationItem {
  id: number;
  name: string;
  address: string;
  latitude: number;
  longitude: number;
  business_hours?: string;
  status: number;
  status_text?: string;
  created_at?: string;
}

/** 驿站表单请求 */
export interface StationFormRequest {
  name: string;
  address: string;
  latitude: number;
  longitude: number;
  business_hours?: string;
  status?: number;
}
