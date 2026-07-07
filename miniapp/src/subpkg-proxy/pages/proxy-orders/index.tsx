import { View, Text } from '@tarojs/components';
import Taro, { usePullDownRefresh, useReachBottom } from '@tarojs/taro';
import { useState, useEffect } from 'react';
import { useProxyStore } from '@/stores/useProxyStore';
import { useUserStore } from '@/stores/useUserStore';
import StatusBadge from '@/components/StatusBadge';
import EmptyState from '@/components/EmptyState';
import { formatAmount, timeAgo } from '@/utils/format';
import './index.scss';

export default function ProxyOrdersPage() {
  const { myOrders, myOrdersLoading, myOrdersHasMore, fetchMyOrders, loadMoreOrders } = useProxyStore();
  const { currentRole } = useUserStore();
  const [roleFilter, setRoleFilter] = useState<'publisher' | 'taker'>('publisher');

  useEffect(() => {
    fetchMyOrders(roleFilter, true);
  }, [roleFilter]);

  usePullDownRefresh(async () => {
    await fetchMyOrders(roleFilter, true);
    Taro.stopPullDownRefresh();
  });

  useReachBottom(() => {
    loadMoreOrders();
  });

  return (
    <View className='proxy-orders'>
      <View className='proxy-orders__tabs'>
        <View
          className={`proxy-orders__tab ${roleFilter === 'publisher' ? 'proxy-orders__tab--active' : ''}`}
          onClick={() => setRoleFilter('publisher')}
        >
          <Text>我发布的</Text>
        </View>
        <View
          className={`proxy-orders__tab ${roleFilter === 'taker' ? 'proxy-orders__tab--active' : ''}`}
          onClick={() => setRoleFilter('taker')}
        >
          <Text>我接的单</Text>
        </View>
      </View>

      <View className='proxy-orders__list'>
        {myOrders.length === 0 && !myOrdersLoading ? (
          <EmptyState title='暂无订单' description={roleFilter === 'publisher' ? '还没有发布过代取任务' : '还没有接过代取任务'} />
        ) : (
          myOrders.map((order) => (
            <View
              key={order.id}
              className='proxy-orders__card'
              onClick={() => Taro.navigateTo({ url: `/subpkg-proxy/pages/proxy-detail/index?id=${order.id}` })}
            >
              <View className='proxy-orders__card-header'>
                <Text className='proxy-orders__reward'>¥{formatAmount(order.reward_amount)}</Text>
                <StatusBadge type='proxy' status={order.status} />
              </View>
              <View className='proxy-orders__card-body'>
                <Text className='proxy-orders__station'>{order.station_name}</Text>
                <Text className='proxy-orders__time'>{timeAgo(order.created_at)}</Text>
              </View>
            </View>
          ))
        )}
      </View>
    </View>
  );
}
