import { View, Text } from '@tarojs/components';
import Taro, { useLoad, useRouter } from '@tarojs/taro';
import { useState, useEffect } from 'react';
import { useParcelStore } from '@/stores/useParcelStore';
import type { PickupCodeInfo } from '@/api/types';
import './index.scss';

export default function PickupCodePage() {
  const router = useRouter();
  const { getPickupCode } = useParcelStore();
  const [codeInfo, setCodeInfo] = useState<PickupCodeInfo | null>(null);

  useEffect(() => {
    const id = router.params.id;
    if (id) {
      getPickupCode(Number(id)).then(setCodeInfo);
    }
    // 页面保持常亮
    Taro.setKeepScreenOn({ keepScreenOn: true });
    return () => {
      Taro.setKeepScreenOn({ keepScreenOn: false });
    };
  }, [router.params.id]);

  if (!codeInfo) {
    return <View className='pickup-code'><Text>加载中...</Text></View>;
  }

  const digits = codeInfo.pickup_code.split('');

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
          <Text className='pickup-code__info-label'>有效期至</Text>
          <Text className='pickup-code__info-value'>{codeInfo.expire_at.split('T')[0]}</Text>
        </View>
      </View>

      <View className='pickup-code__tips'>
        <Text className='pickup-code__tip'>请出示此码给驿站工作人员核验</Text>
        <Text className='pickup-code__tip'>出库后取件码自动失效</Text>
      </View>
    </View>
  );
}
