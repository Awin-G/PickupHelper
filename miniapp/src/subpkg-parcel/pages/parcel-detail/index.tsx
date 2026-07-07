import { View, Text } from '@tarojs/components';
import Taro, { useLoad, useRouter } from '@tarojs/taro';
import { useState, useEffect } from 'react';
import { Button } from '@nutui/nutui-react-taro';
import { useParcelStore } from '@/stores/useParcelStore';
import StatusBadge from '@/components/StatusBadge';
import { formatTime, timeAgo } from '@/utils/format';
import type { Parcel } from '@/api/types';
import './index.scss';

export default function ParcelDetailPage() {
  const router = useRouter();
  const { getParcelDetail } = useParcelStore();
  const [parcel, setParcel] = useState<Parcel | null>(null);

  useEffect(() => {
    const id = router.params.id;
    if (id) {
      getParcelDetail(Number(id)).then(setParcel);
    }
  }, [router.params.id]);

  if (!parcel) {
    return <View className='parcel-detail'><Text>加载中...</Text></View>;
  }

  const isPending = parcel.status === 1 || parcel.status === 3;

  return (
    <View className='parcel-detail'>
      <View className='parcel-detail__header'>
        <StatusBadge type='parcel' status={parcel.status} />
      </View>

      <View className='parcel-detail__card'>
        <View className='parcel-detail__row'>
          <Text className='parcel-detail__label'>快递公司</Text>
          <Text className='parcel-detail__value'>{parcel.courier_company}</Text>
        </View>
        <View className='parcel-detail__row'>
          <Text className='parcel-detail__label'>快递单号</Text>
          <Text className='parcel-detail__value'>{parcel.tracking_no}</Text>
        </View>
        <View className='parcel-detail__row'>
          <Text className='parcel-detail__label'>货架编号</Text>
          <Text className='parcel-detail__value'>{parcel.shelf_code}</Text>
        </View>
        <View className='parcel-detail__row'>
          <Text className='parcel-detail__label'>入库时间</Text>
          <Text className='parcel-detail__value'>{formatTime(parcel.storage_time)}</Text>
        </View>
        {parcel.remarks && (
          <View className='parcel-detail__row'>
            <Text className='parcel-detail__label'>备注</Text>
            <Text className='parcel-detail__value'>{parcel.remarks}</Text>
          </View>
        )}
        {parcel.is_fragile && (
          <View className='parcel-detail__row'>
            <Text className='parcel-detail__label'>标记</Text>
            <Text className='parcel-detail__value parcel-detail__value--warning'>易碎品</Text>
          </View>
        )}
      </View>

      {isPending && (
        <View className='parcel-detail__actions'>
          <Button
            type='primary'
            block
            className='parcel-detail__btn'
            onClick={() => Taro.navigateTo({ url: `/subpkg-parcel/pages/pickup-code/index?id=${parcel.id}` })}
          >
            查看取件码
          </Button>
          <Button
            block
            className='parcel-detail__btn parcel-detail__btn--outline'
            onClick={() => Taro.navigateTo({ url: `/subpkg-proxy/pages/proxy-publish/index?parcelId=${parcel.id}` })}
          >
            找人代取
          </Button>
          <Button
            block
            className='parcel-detail__btn parcel-detail__btn--outline'
            onClick={() => Taro.navigateTo({ url: '/subpkg-user/pages/station-map/index' })}
          >
            导航到驿站
          </Button>
        </View>
      )}
    </View>
  );
}
