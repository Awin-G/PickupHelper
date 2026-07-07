import { PropsWithChildren } from 'react';
import { useLaunch } from '@tarojs/taro';
import { useUserStore } from '@/stores/useUserStore';
import { useParcelStore } from '@/stores/useParcelStore';
import { useNotificationStore } from '@/stores/useNotificationStore';

import './app.scss';

function App({ children }: PropsWithChildren<any>) {
  useLaunch(async () => {
    const { isLoggedIn, refreshAuth, fetchUserInfo } = useUserStore.getState();

    if (isLoggedIn) {
      // 尝试刷新 Token 恢复登录态
      const refreshed = await refreshAuth();
      if (refreshed) {
        // 并行拉取用户信息和待取件数
        await Promise.all([
          fetchUserInfo(),
          useParcelStore.getState().fetchPendingCount(),
          useNotificationStore.getState().fetchUnreadCount(),
        ]);
      }
    }
  });

  return children;
}

export default App;
