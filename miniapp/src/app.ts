import { PropsWithChildren } from 'react';
import { useLaunch } from '@tarojs/taro';
import { useUserStore } from '@/stores/useUserStore';
import { useParcelStore } from '@/stores/useParcelStore';
import { useNotificationStore } from '@/stores/useNotificationStore';
import { storage } from '@/utils/storage';

import './app.scss';

const IS_DEV = process.env.NODE_ENV === 'development';

function App({ children }: PropsWithChildren<any>) {
  useLaunch(async () => {
    const { isLoggedIn, refreshAuth, fetchUserInfo, login } = useUserStore.getState();

    // 开发环境自动 mock 登录
    if (IS_DEV && !isLoggedIn) {
      storage.set('token', 'mock_token_xxx');
      storage.set('refresh_token', 'mock_refresh_xxx');
      storage.set('currentRole', 'receiver');
      useUserStore.setState({
        token: 'mock_token_xxx',
        refreshToken: 'mock_refresh_xxx',
        isLoggedIn: true,
        currentRole: 'receiver',
      });
    }

    if (isLoggedIn || IS_DEV) {
      await Promise.all([
        fetchUserInfo(),
        useParcelStore.getState().fetchPendingCount(),
        useNotificationStore.getState().fetchUnreadCount(),
      ]);
    }
  });

  return children;
}

export default App;
