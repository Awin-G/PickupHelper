/** 包裹状态 */
export const PARCEL_STATUS = {
  PENDING: 1,    // 待取
  PICKED: 2,     // 已取
  STRANDED: 3,   // 滞留
  RETURNED: 4,   // 已退件
  ABNORMAL: 5,   // 异常
} as const;

export const PARCEL_STATUS_MAP: Record<number, { text: string; color: string }> = {
  1: { text: '待取', color: '#52C41A' },
  2: { text: '已取', color: '#999999' },
  3: { text: '滞留', color: '#FF4D4F' },
  4: { text: '已退件', color: '#999999' },
  5: { text: '异常', color: '#FAAD14' },
};

/** 代取订单状态 */
export const PROXY_STATUS = {
  PENDING: 1,       // 待接单
  DELIVERING: 2,    // 配送中
  CONFIRMING: 3,    // 待确认
  COMPLETED: 4,     // 已完成
  CANCELLED: 5,     // 已取消
  FAILED: 6,        // 取件失败
} as const;

export const PROXY_STATUS_MAP: Record<number, { text: string; color: string }> = {
  1: { text: '待接单', color: '#1890FF' },
  2: { text: '配送中', color: '#FAAD14' },
  3: { text: '待确认', color: '#722ED1' },
  4: { text: '已完成', color: '#52C41A' },
  5: { text: '已取消', color: '#999999' },
  6: { text: '取件失败', color: '#FF4D4F' },
};

/** 用户角色 */
export const USER_ROLE = {
  RECEIVER: 'receiver',  // 收件人
  RUNNER: 'runner',      // 跑腿员
} as const;

/** 跑腿员审核状态 */
export const RUNNER_STATUS = {
  NOT_APPLIED: 0,  // 未申请
  PENDING: 1,      // 审核中
  APPROVED: 2,     // 已通过
  REJECTED: 3,     // 已拒绝
} as const;

/** 消息类型 */
export const NOTIFICATION_TYPE = {
  INBOUND: 1,      // 入库通知
  REMINDER: 2,     // 滞留催取
  PROXY: 3,        // 代取状态通知
  SYSTEM: 4,       // 系统通知
} as const;

export const NOTIFICATION_TYPE_MAP: Record<number, { icon: string; title: string }> = {
  1: { icon: '📦', title: '入库通知' },
  2: { icon: '⚠️', title: '滞留催取' },
  3: { icon: '🚶', title: '代取状态通知' },
  4: { icon: 'ℹ️', title: '系统通知' },
};

/** 推荐悬赏金额 */
export const RECOMMENDED_REWARDS = [
  { label: '普通', value: 3 },
  { label: '优先', value: 5 },
  { label: '加急', value: 8 },
] as const;

/** API 错误码 */
export const API_ERROR_CODE = {
  SUCCESS: 0,
  TOKEN_EXPIRED: 10002,
  USER_BLACKLISTED: 20001,
} as const;
