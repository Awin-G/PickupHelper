import { View, Text } from '@tarojs/components';
import Taro, { usePullDownRefresh, useReachBottom } from '@tarojs/taro';
import { useState, useEffect } from 'react';
import { useProxyStore } from '@/stores/useProxyStore';
import { useUserStore } from '@/stores/useUserStore';
import StatusBadge from '@/components/StatusBadge';
import EmptyState from '@/components/EmptyState';
import { formatAmount, timeAgo } from '@/utils/format';
import './index.scss';

export default function OrdersPage() {
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
    if (myOrdersHasMore && !myOrdersLoading) {
      loadMoreOrders();
    }
  });

  return (
    <View className='orders-page'>
      <View className='orders-page__tabs'>
        <View
          className={`orders-page__tab ${roleFilter === 'publisher' ? 'orders-page__tab--active' : ''}`}
          onClick={() => setRoleFilter('publisher')}
        >
          <Text>我发布的</Text>
        </View>
        <View
          className={`orders-page__tab ${roleFilter === 'taker' ? 'orders-page__tab--active' : ''}`}
          onClick={() => setRoleFilter('taker')}
        >
          <Text>我接的单</Text>
        </View>
      </View>

      <View className='orders-page__list'>
        {myOrders.length === 0 && !myOrdersLoading ? (
          <EmptyState title='暂无订单' description={roleFilter === 'publisher' ? '还没有发布过代取任务' : '还没有接过代取任务'} />
        ) : (
          myOrders.map((order) => (
            <View
              key={order.id}
              className='orders-page__card'
              onClick={() => Taro.navigateTo({ url: `/subpkg-proxy/pages/proxy-detail/index?id=${order.id}` })}
            >
              <View className='orders-page__card-header'>
                <Text className='orders-page__reward'>¥{formatAmount(order.reward_amount)}</Text>
                <StatusBadge type='proxy' status={order.status} />
              </View>
              <View className='orders-page__card-body'>
                <Text className='orders-page__station'>{order.station_name}</Text>
                <Text className='orders-page__time'>{timeAgo(order.created_at)}</Text>
              </View>
            </View>
          ))
        )}
      </View>
    </View>
  );
}
