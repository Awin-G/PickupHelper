/** 手机号校验 */
export function isValidPhone(phone: string): boolean {
  return /^1[3-9]\d{9}$/.test(phone);
}

/** 验证码校验 (4-6位数字) */
export function isValidCode(code: string): boolean {
  return /^\d{4,6}$/.test(code);
}

/** 金额校验 (0.01 ~ 500.00) */
export function isValidAmount(amount: number | string): boolean {
  const val = typeof amount === 'string' ? parseFloat(amount) : amount;
  return !isNaN(val) && val >= 0.01 && val <= 500;
}

/** 非空校验 */
export function isNotEmpty(value: string | undefined | null): boolean {
  return !!value && value.trim().length > 0;
}

/** 字符串长度校验 */
export function isLengthInRange(value: string, min: number, max: number): boolean {
  return value.length >= min && value.length <= max;
}
