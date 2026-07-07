import { useState, useCallback } from 'react';
import Taro from '@tarojs/taro';

interface Location {
  latitude: number;
  longitude: number;
}

type AuthStatus = 'authorized' | 'denied' | 'undetermined';

export function useGeoLocation() {
  const [location, setLocation] = useState<Location | null>(null);
  const [authStatus, setAuthStatus] = useState<AuthStatus>('undetermined');

  const requestLocation = useCallback(async () => {
    try {
      const res = await Taro.getLocation({ type: 'gcj02' });
      setLocation({ latitude: res.latitude, longitude: res.longitude });
      setAuthStatus('authorized');
      return { latitude: res.latitude, longitude: res.longitude };
    } catch {
      setAuthStatus('denied');
      return null;
    }
  }, []);

  const ensureLocation = useCallback(async (): Promise<Location | null> => {
    try {
      const setting = await Taro.getSetting();
      if (setting.authSetting['scope.userLocation'] === false) {
        setAuthStatus('denied');
        return null;
      }
      if (!location) {
        return await requestLocation();
      }
      return location;
    } catch {
      return await requestLocation();
    }
  }, [location, requestLocation]);

  const openSetting = useCallback(async () => {
    try {
      await Taro.openSetting();
      const setting = await Taro.getSetting();
      if (setting.authSetting['scope.userLocation'] !== false) {
        setAuthStatus('authorized');
        await requestLocation();
      }
    } catch {
      // 用户取消
    }
  }, [requestLocation]);

  return {
    location,
    authStatus,
    requestLocation,
    ensureLocation,
    openSetting,
  };
}
