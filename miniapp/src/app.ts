import { PropsWithChildren } from 'react';
import { useLaunch } from '@tarojs/taro';
import { useUserStore } from '@/stores/useUserStore';
import { useParcelStore } from '@/stores/useParcelStore';
import { useNotificationStore } from '@/stores/useNotificationStore';
import { storage } from '@/utils/storage';

import './app.scss';

// 检查是否使用 mock
let USE_MOCK = true;
if (typeof window !== 'undefined') {
  const urlParams = new URLSearchParams(window.location.search);
  const mockParam = urlParams.get('mock');
  if (mockParam === 'false') {
    USE_MOCK = false;
  }
  const storedMock = localStorage.getItem('pickup_use_mock');
  if (storedMock === 'false') {
    USE_MOCK = false;
  }
}

function App({ children }: PropsWithChildren<any>) {
  useLaunch(async () => {
    const { isLoggedIn, fetchUserInfo } = useUserStore.getState();

    // 仅 mock 模式自动登录
    if (USE_MOCK && !isLoggedIn) {
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

    // 已登录时拉取数据
    if (isLoggedIn || USE_MOCK) {
      await fetchUserInfo();
      // 先拉取包裹列表，再计算待取件数
      await useParcelStore.getState().fetchMyParcels(true);
      useParcelStore.getState().fetchPendingCount();
      useNotificationStore.getState().fetchUnreadCount();
    }
  });

  return children;
}

export default App;
