import { View, Text } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { useState, useEffect } from 'react';
import { Button } from '@nutui/nutui-react-taro';
import { stationApi } from '@/api/stations';
import EmptyState from '@/components/EmptyState';
import type { Station } from '@/api/types';
import './index.scss';

export default function StationMapPage() {
  const [stations, setStations] = useState<Station[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    stationApi.list()
      .then(setStations)
      .catch((err: any) => Taro.showToast({ title: err.msg || err.message || '获取驿站列表失败', icon: 'none', duration: 3000 }))
      .finally(() => setLoading(false));
  }, []);

  const handleNavigate = (station: Station) => {
    Taro.showModal({
      title: station.name,
      content: `地址：${station.address}${station.hours ? '\n营业时间：' + station.hours : ''}`,
      confirmText: '导航',
      success: (res) => {
        if (res.confirm) {
          Taro.showToast({ title: '正在打开地图...', icon: 'success' });
        }
      },
    });
  };

  if (!loading && stations.length === 0) {
    return <EmptyState title='暂无驿站信息' />;
  }

  return (
    <View className='station-map'>
      <View className='station-map__list'>
        {stations.map((s) => (
          <View key={s.id} className='station-map__card'>
            <View className='station-map__card-header'>
              <Text className='station-map__name'>{s.name}</Text>
              {s.distance && <Text className='station-map__distance'>{s.distance}</Text>}
            </View>
            <Text className='station-map__address'>{s.address}</Text>
            {s.hours && <Text className='station-map__hours'>营业时间: {s.hours}</Text>}
            <View className='station-map__card-footer'>
              <Button
                type='primary'
                size='small'
                className='station-map__nav-btn'
                onClick={() => handleNavigate(s)}
              >
                导航
              </Button>
            </View>
          </View>
        ))}
      </View>
    </View>
  );
}
