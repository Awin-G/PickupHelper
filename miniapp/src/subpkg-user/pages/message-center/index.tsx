import { View, Text } from '@tarojs/components';
import Taro, { usePullDownRefresh, useReachBottom } from '@tarojs/taro';
import { useState, useEffect } from 'react';
import { useNotificationStore } from '@/stores/useNotificationStore';
import EmptyState from '@/components/EmptyState';
import { timeAgo } from '@/utils/format';
import { NOTIFICATION_TYPE_MAP } from '@/utils/constants';
import './index.scss';

export default function MessageCenterPage() {
  const {
    notifications, loading, hasMore,
    fetchNotifications, loadMore, markAsRead, markAllAsRead, fetchUnreadCount,
  } = useNotificationStore();

  useEffect(() => {
    fetchNotifications(true);
  }, []);

  usePullDownRefresh(async () => {
    await fetchNotifications(true);
    Taro.stopPullDownRefresh();
  });

  useReachBottom(() => {
    loadMore();
  });

  const handleMarkAllRead = async () => {
    await markAllAsRead();
    Taro.showToast({ title: '全部已读', icon: 'success' });
  };

  const handleTapNotification = async (id: number, parcelId: number | null) => {
    await markAsRead(id);
    fetchUnreadCount();
    if (parcelId) {
      Taro.navigateTo({ url: `/subpkg-parcel/pages/parcel-detail/index?id=${parcelId}` });
    }
  };

  return (
    <View className='message-center'>
      <View className='message-center__header'>
        <View className='message-center__mark-all' onClick={handleMarkAllRead}>
          <Text>全部已读</Text>
        </View>
      </View>

      <View className='message-center__list'>
        {notifications.length === 0 && !loading ? (
          <EmptyState title='暂无消息' description='暂时没有新消息' />
        ) : (
          notifications.map((n) => {
            const typeConfig = NOTIFICATION_TYPE_MAP[n.type] || { icon: 'ℹ️', title: '通知' };
            return (
              <View
                key={n.id}
                className={`message-center__item ${n.is_read ? 'message-center__item--read' : ''}`}
                onClick={() => handleTapNotification(n.id, n.parcel_id)}
              >
                <Text className='message-center__icon'>{typeConfig.icon}</Text>
                <View className='message-center__content'>
                  <Text className='message-center__title'>{typeConfig.title}</Text>
                  <Text className='message-center__text'>{n.content}</Text>
                  <Text className='message-center__time'>{timeAgo(n.created_at)}</Text>
                </View>
                {!n.is_read && <View className='message-center__dot' />}
              </View>
            );
          })
        )}
      </View>
    </View>
  );
}
