import { useCallback } from 'react';
import Taro from '@tarojs/taro';

/** 配合页面 onPullDownRefresh 使用 */
export function usePullRefresh(onRefresh: () => Promise<void>) {
  const handlePullRefresh = useCallback(async () => {
    try {
      await onRefresh();
    } finally {
      Taro.stopPullDownRefresh();
    }
  }, [onRefresh]);

  return { handlePullRefresh };
}
