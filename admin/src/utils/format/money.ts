/**
 * 格式化金额：5.00
 */
export function formatMoney(amount: number | null | undefined): string {
  if (amount === null || amount === undefined) return "0.00";
  return amount.toFixed(2);
}

/**
 * 格式化金额带符号：¥5.00
 */
export function formatMoneyWithSymbol(
  amount: number | null | undefined
): string {
  return `¥${formatMoney(amount)}`;
}
