import { View, Text } from '@tarojs/components';
import './index.scss';

interface SkeletonProps {
  rows?: number;
  type?: 'card' | 'list' | 'stats';
}

export default function Skeleton({ rows = 3, type = 'card' }: SkeletonProps) {
  if (type === 'stats') {
    return (
      <View className='skeleton skeleton--stats'>
        <View className='skeleton__stat-item'>
          <View className='skeleton__circle' />
          <View className='skeleton__line skeleton__line--short' />
        </View>
      </View>
    );
  }

  if (type === 'list') {
    return (
      <View className='skeleton skeleton--list'>
        {Array.from({ length: rows }).map((_, i) => (
          <View key={i} className='skeleton__list-item'>
            <View className='skeleton__line' />
            <View className='skeleton__line skeleton__line--short' />
          </View>
        ))}
      </View>
    );
  }

  return (
    <View className='skeleton skeleton--card'>
      {Array.from({ length: rows }).map((_, i) => (
        <View key={i} className='skeleton__card'>
          <View className='skeleton__card-header'>
            <View className='skeleton__line' />
            <View className='skeleton__badge' />
          </View>
          <View className='skeleton__card-body'>
            <View className='skeleton__line' />
            <View className='skeleton__line skeleton__line--short' />
            <View className='skeleton__line skeleton__line--medium' />
          </View>
        </View>
      ))}
    </View>
  );
}
