import { View, Text } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { useState } from 'react';
import { Button } from '@nutui/nutui-react-taro';
import './index.scss';

const MOCK_STATIONS = [
  { id: 1, name: '南门菜鸟驿站', address: '学校南门左侧50米', distance: '350m', hours: '08:00-22:00' },
  { id: 2, name: '北门驿站', address: '学校北门右侧100米', distance: '800m', hours: '09:00-21:00' },
  { id: 3, name: '东门快递点', address: '学校东门对面', distance: '1.2km', hours: '08:30-20:00' },
];

export default function StationMapPage() {
  const [stations] = useState(MOCK_STATIONS);

  const handleNavigate = (station: typeof MOCK_STATIONS[0]) => {
    Taro.showModal({
      title: station.name,
      content: `地址：${station.address}\n营业时间：${station.hours}`,
      confirmText: '导航',
      success: (res) => {
        if (res.confirm) {
          Taro.showToast({ title: '正在打开地图...', icon: 'success' });
        }
      },
    });
  };

  return (
    <View className='station-map'>
      <View className='station-map__list'>
        {stations.map((s) => (
          <View key={s.id} className='station-map__card'>
            <View className='station-map__card-header'>
              <Text className='station-map__name'>{s.name}</Text>
              <Text className='station-map__distance'>{s.distance}</Text>
            </View>
            <Text className='station-map__address'>{s.address}</Text>
            <Text className='station-map__hours'>营业时间: {s.hours}</Text>
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
