import { View, Text } from '@tarojs/components';
import { formatTime, timeAgo } from '@/utils/format';
import { PARCEL_STATUS_MAP } from '@/utils/constants';
import StatusBadge from '../StatusBadge';
import type { Parcel } from '@/api/types';
import './index.scss';

interface ParcelCardProps {
  parcel: Parcel;
  showPickupCode?: boolean;
  onTap?: (parcel: Parcel) => void;
  onPickup?: (parcel: Parcel) => void;
  onProxy?: (parcel: Parcel) => void;
}

export default function ParcelCard({
  parcel,
  showPickupCode = false,
  onTap,
  onPickup,
  onProxy,
}: ParcelCardProps) {
  const isPending = parcel.status === 1 || parcel.status === 3;

  return (
    <View className='parcel-card' onClick={() => onTap?.(parcel)}>
      <View className='parcel-card__header'>
        <Text className='parcel-card__company'>{parcel.courier_company}</Text>
        <StatusBadge type='parcel' status={parcel.status} />
      </View>

      <View className='parcel-card__body'>
        <View className='parcel-card__row'>
          <Text className='parcel-card__label'>货架</Text>
          <Text className='parcel-card__value'>{parcel.shelf_code}</Text>
        </View>
        <View className='parcel-card__row'>
          <Text className='parcel-card__label'>单号</Text>
          <Text className='parcel-card__value'>{parcel.tracking_no}</Text>
        </View>
        <View className='parcel-card__row'>
          <Text className='parcel-card__label'>入库</Text>
          <Text className='parcel-card__value'>{timeAgo(parcel.storage_time)}</Text>
        </View>
      </View>

      {isPending && (
        <View className='parcel-card__footer'>
          {showPickupCode && (
            <View
              className='parcel-card__btn parcel-card__btn--primary'
              onClick={(e) => {
                e.stopPropagation();
                onPickup?.(parcel);
              }}
            >
              <Text>查看取件码</Text>
            </View>
          )}
          <View
            className='parcel-card__btn parcel-card__btn--default'
            onClick={(e) => {
              e.stopPropagation();
              onProxy?.(parcel);
            }}
          >
            <Text>找人代取</Text>
          </View>
        </View>
      )}
    </View>
  );
}
