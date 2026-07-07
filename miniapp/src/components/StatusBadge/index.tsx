import { View, Text } from '@tarojs/components';
import { PARCEL_STATUS_MAP, PROXY_STATUS_MAP } from '@/utils/constants';
import './index.scss';

interface StatusBadgeProps {
  type: 'parcel' | 'proxy';
  status: number;
}

export default function StatusBadge({ type, status }: StatusBadgeProps) {
  const statusMap = type === 'parcel' ? PARCEL_STATUS_MAP : PROXY_STATUS_MAP;
  const config = statusMap[status];

  if (!config) return null;

  return (
    <View className='status-badge' style={{ backgroundColor: config.color + '15', color: config.color }}>
      <Text className='status-badge__text'>{config.text}</Text>
    </View>
  );
}
