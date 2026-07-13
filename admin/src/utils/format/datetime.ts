import dayjs from "dayjs";

const DEFAULT_FORMAT = "YYYY-MM-DD HH:mm:ss";
const DATE_FORMAT = "YYYY-MM-DD";

/**
 * 格式化日期时间
 */
export function formatDateTime(date: string | Date | null | undefined): string {
  if (!date) return "-";
  return dayjs(date).format(DEFAULT_FORMAT);
}

/**
 * 格式化日期
 */
export function formatDate(date: string | Date | null | undefined): string {
  if (!date) return "-";
  return dayjs(date).format(DATE_FORMAT);
}
