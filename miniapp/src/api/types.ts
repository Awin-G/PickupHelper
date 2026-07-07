/** 通用分页响应 */
export interface PaginatedList<T> {
  list: T[];
  total: number;
  page: number;
  page_size: number;
}

/** 通用 API 响应 */
export interface ApiResponse<T> {
  code: number;
  msg: string;
  data: T;
  trace_id: string;
}

/** 用户信息 */
export interface UserInfo {
  id: number;
  phone: string;
  nickname: string;
  avatar: string;
  user_type: number;       // 1-收件人, 2-跑腿员
  runner_status: number;   // 0-未申请, 1-审核中, 2-已通过, 3-已拒绝
  credit_score: number;
  is_blacklisted: boolean;
}

/** 包裹 */
export interface Parcel {
  id: number;
  station_id: number;
  tracking_no: string;
  courier_company: string;
  shelf_code: string;
  pickup_code?: string;
  receiver_phone: string;
  receiver_name: string;
  status: number;
  status_text: string;
  weight: number;
  is_fragile: boolean;
  remarks: string;
  storage_time: string;
  pickup_time: string | null;
  return_time: string | null;
  notify_count: number;
}

/** 取件码信息 */
export interface PickupCodeInfo {
  pickup_code: string;
  qr_url: string;
  expire_at: string;
}

/** 代取任务（大厅列表项） */
export interface ProxyTask {
  id: number;
  parcel_id: number;
  station_id: number;
  station_name: string;
  reward_amount: number;
  deadline: string;
  remark: string;
  created_at: string;
}

/** 代取订单 */
export interface ProxyOrder {
  id: number;
  parcel_id: number;
  station_name: string;
  publisher_id: number;
  publisher_nickname: string;
  taker_id: number | null;
  taker_nickname: string | null;
  reward_amount: number;
  status: number;
  status_text: string;
  temp_pickup_code?: string;
  deadline: string;
  delivery_time: string | null;
  created_at: string;
}

/** 通知 */
export interface Notification {
  id: number;
  title: string;
  content: string;
  type: number;
  is_read: boolean;
  parcel_id: number | null;
  created_at: string;
}

/** 任务查询参数 */
export interface TaskQueryParams {
  page?: number;
  page_size?: number;
  sort_by?: 'reward' | 'created_at' | 'deadline';
  station_id?: number;
}

/** 发布代取参数 */
export interface PublishParams {
  parcel_id: number;
  reward_amount: number;
  deadline: string;
  remark?: string;
}

/** 登录参数 */
export interface LoginParams {
  phone: string;
  code: string;
}

/** 登录响应 */
export interface LoginResult {
  token: string;
  refresh_token: string;
  user: UserInfo;
}

/** 业务错误 */
export class BusinessError extends Error {
  code: number;
  msg: string;

  constructor(code: number, msg: string) {
    super(msg);
    this.name = 'BusinessError';
    this.code = code;
    this.msg = msg;
  }
}
