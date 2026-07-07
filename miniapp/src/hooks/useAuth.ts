import { useEffect } from 'react';
import Taro from '@tarojs/taro';
import { useUserStore } from '@/stores/useUserStore';

/** 登录态守卫：未登录时跳转登录页 */
export function useAuth(redirect = true) {
  const { isLoggedIn, token } = useUserStore();

  useEffect(() => {
    if (!isLoggedIn && redirect) {
      Taro.redirectTo({ url: '/pages/login/index' });
    }
  }, [isLoggedIn, redirect]);

  return { isLoggedIn, token };
}

/** 需要登录才能执行的操作 */
export function useAuthGuard() {
  const { isLoggedIn } = useUserStore();

  const withAuth = (callback: () => void) => {
    if (!isLoggedIn) {
      Taro.navigateTo({ url: '/pages/login/index' });
      return;
    }
    callback();
  };

  return { withAuth, isLoggedIn };
}
