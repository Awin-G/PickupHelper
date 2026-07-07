/** 格式化手机号为脱敏形式: 138****0000 */
export function maskPhone(phone: string): string {
  if (!phone || phone.length !== 11) return phone;
  return phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2');
}

/** 格式化金额: 5 -> "5.00" */
export function formatAmount(amount: number): string {
  return amount.toFixed(2);
}

/** 格式化时间: "2026-07-05T10:30:00Z" -> "07-05 10:30" */
export function formatTime(time: string, format: 'datetime' | 'date' | 'time' = 'datetime'): string {
  if (!time) return '';
  const d = new Date(time);
  const pad = (n: number) => String(n).padStart(2, '0');
  const MM = pad(d.getMonth() + 1);
  const DD = pad(d.getDate());
  const hh = pad(d.getHours());
  const mm = pad(d.getMinutes());

  if (format === 'date') return `${d.getFullYear()}-${MM}-${DD}`;
  if (format === 'time') return `${hh}:${mm}`;
  return `${MM}-${DD} ${hh}:${mm}`;
}

/** 相对时间: "2分钟前", "3天前" */
export function timeAgo(time: string): string {
  if (!time) return '';
  const now = Date.now();
  const diff = now - new Date(time).getTime();
  const minute = 60 * 1000;
  const hour = 60 * minute;
  const day = 24 * hour;

  if (diff < minute) return '刚刚';
  if (diff < hour) return `${Math.floor(diff / minute)}分钟前`;
  if (diff < day) return `${Math.floor(diff / hour)}小时前`;
  if (diff < 30 * day) return `${Math.floor(diff / day)}天前`;
  return formatTime(time, 'date');
}

/** 格式化快递单号: 隐藏中间部分 */
export function maskTrackingNo(trackingNo: string): string {
  if (!trackingNo || trackingNo.length <= 6) return trackingNo;
  return trackingNo.slice(0, 3) + '***' + trackingNo.slice(-3);
}

/** 生成随机昵称: 用户 + 4位随机数 */
export function generateNickname(): string {
  const randomNum = Math.floor(1000 + Math.random() * 9000);
  return `用户${randomNum}`;
}
