import { View, Text, Canvas } from '@tarojs/components';
import Taro, { useRouter, useDidShow } from '@tarojs/taro';
import { useState, useEffect, useRef } from 'react';
import { useParcelStore } from '@/stores/useParcelStore';
import QRCode from 'qrcode-generator';
import type { Parcel } from '@/api/types';
import './index.scss';

export default function PickupCodePage() {
  const router = useRouter();
  const { getParcelDetail } = useParcelStore();
  const [parcel, setParcel] = useState<Parcel | null>(null);
  const [error, setError] = useState('');
  const canvasReady = useRef(false);

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
        .catch((err: any) => setError(err.msg || err.message || '加载失败'));
    }
    Taro.setKeepScreenOn({ keepScreenOn: true });
    return () => {
      Taro.setKeepScreenOn({ keepScreenOn: false });
    };
  }, [router.params.id]);

  const drawQR = (code: string) => {
    const query = Taro.createSelectorQuery();
    query.select('#pickupCodeQR')
      .fields({ node: true, size: true })
      .exec((res) => {
        if (!res || !res[0]) return;
        const canvas = res[0].node;
        const ctx = canvas.getContext('2d');
        const dpr = Taro.getSystemInfoSync().pixelRatio;
        const size = 300;
        canvas.width = size * dpr;
        canvas.height = size * dpr;
        ctx.scale(dpr, dpr);

        const qr = QRCode(0, 'M');
        qr.addData(code);
        qr.make();
        const moduleCount = qr.getModuleCount();
        const cellSize = Math.floor(size / moduleCount);
        const totalSize = cellSize * moduleCount;
        const offset = Math.floor((size - totalSize) / 2);

        ctx.fillStyle = '#ffffff';
        ctx.fillRect(0, 0, size, size);

        ctx.fillStyle = '#000000';
        for (let r = 0; r < moduleCount; r++) {
          for (let c = 0; c < moduleCount; c++) {
            if (qr.isDark(r, c)) {
              ctx.fillRect(offset + c * cellSize, offset + r * cellSize, cellSize, cellSize);
            }
          }
        }
      });
  };

  useDidShow(() => {
    if (canvasReady.current && parcel && parcel.pickup_code) {
      setTimeout(() => drawQR(parcel.pickup_code), 100);
    }
  });

  useEffect(() => {
    if (parcel && parcel.pickup_code) {
      canvasReady.current = true;
      setTimeout(() => drawQR(parcel.pickup_code), 500);
    }
  }, [parcel]);

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

      <Canvas
        className='pickup-code__qr'
        id='pickupCodeQR'
        type='2d'
        style='width: 600rpx; height: 600rpx;'
      />

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
