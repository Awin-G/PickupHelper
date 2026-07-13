import { PropsWithChildren } from 'react';
import Taro, { useLaunch } from '@tarojs/taro';
import { useUserStore } from '@/stores/useUserStore';
import { useParcelStore } from '@/stores/useParcelStore';
import { useNotificationStore } from '@/stores/useNotificationStore';
import { storage } from '@/utils/storage';

import './app.scss';

function App({ children }: PropsWithChildren<any>) {
  useLaunch(async () => {
    const { isLoggedIn, fetchUserInfo } = useUserStore.getState();

    if (!isLoggedIn) {
      Taro.reLaunch({ url: '/pages/login/index' });
      return;
    }

    // 已登录时拉取数据
    await fetchUserInfo();
    await useParcelStore.getState().fetchMyParcels(true);
    useParcelStore.getState().fetchPendingCount();
    useNotificationStore.getState().fetchUnreadCount();
  });

  return children;
}

export default App;
