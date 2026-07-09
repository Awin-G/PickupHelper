import { View, Text } from '@tarojs/components';
import Taro, { useRouter } from '@tarojs/taro';
import { useState, useEffect } from 'react';
import { useParcelStore } from '@/stores/useParcelStore';
import type { Parcel } from '@/api/types';
import './index.scss';

export default function PickupCodePage() {
  const router = useRouter();
  const { getParcelDetail } = useParcelStore();
  const [parcel, setParcel] = useState<Parcel | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    const id = router.params.id;
    if (id) {
      getParcelDetail(Number(id))
        .then((p) => {
          if (p.pickup_code) {
            setParcel(p);
          } else {
            setError('无法获取取件码');
          }
        })
        .catch(() => setError('加载失败'));
    }
    Taro.setKeepScreenOn({ keepScreenOn: true });
    return () => {
      Taro.setKeepScreenOn({ keepScreenOn: false });
    };
  }, [router.params.id]);

  if (error) {
    return (
      <View className='pickup-code'>
        <View className='pickup-code__error'>
          <Text>{error}</Text>
        </View>
      </View>
    );
  }

  if (!parcel || !parcel.pickup_code) {
    return (
      <View className='pickup-code'>
        <View className='pickup-code__loading'>
          <Text>加载中...</Text>
        </View>
      </View>
    );
  }

  const digits = parcel.pickup_code.split('');

  return (
    <View className='pickup-code'>
      <View className='pickup-code__title'>
        <Text>取件码</Text>
      </View>

      <View className='pickup-code__qr'>
        <View className='pickup-code__qr-placeholder'>
          <Text className='pickup-code__qr-text'>QR</Text>
        </View>
      </View>

      <View className='pickup-code__digits'>
        {digits.map((d, i) => (
          <View key={i} className='pickup-code__digit'>
            <Text>{d}</Text>
          </View>
        ))}
      </View>

      <View className='pickup-code__info'>
        <View className='pickup-code__info-row'>
          <Text className='pickup-code__info-label'>快递公司</Text>
          <Text className='pickup-code__info-value'>{parcel.courier_company}</Text>
        </View>
        <View className='pickup-code__info-row'>
          <Text className='pickup-code__info-label'>货架编号</Text>
          <Text className='pickup-code__info-value'>{parcel.shelf_code}</Text>
        </View>
      </View>

      <View className='pickup-code__tips'>
        <Text className='pickup-code__tip'>请出示此码给驿站工作人员核验</Text>
        <Text className='pickup-code__tip'>出库后取件码自动失效</Text>
      </View>
    </View>
  );
}
