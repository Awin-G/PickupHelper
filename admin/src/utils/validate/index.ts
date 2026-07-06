/**
 * 验证手机号格式
 */
export function isValidPhone(phone: string): boolean {
  return /^1[3-9]\d{9}$/.test(phone);
}

/**
 * 验证取件码格式（6位数字）
 */
export function isValidPickupCode(code: string): boolean {
  return /^\d{6}$/.test(code);
}
