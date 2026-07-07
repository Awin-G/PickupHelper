import { View, Text } from '@tarojs/components';
import Taro, { useLoad, usePullDownRefresh, useReachBottom } from '@tarojs/taro';
import { useUserStore } from '@/stores/useUserStore';
import { useParcelStore } from '@/stores/useParcelStore';
import { useProxyStore } from '@/stores/useProxyStore';
import { useNotificationStore } from '@/stores/useNotificationStore';
import ParcelCard from '@/components/ParcelCard';
import EmptyState from '@/components/EmptyState';
import { formatAmount, timeAgo } from '@/utils/format';
import './index.scss';

export default function Index() {
  const { isLoggedIn, currentRole, userInfo } = useUserStore();
  const { myParcels, pendingCount, loading, fetchMyParcels, loadMore } = useParcelStore();
  const { taskList, fetchTasks } = useProxyStore();
  const { unreadCount } = useNotificationStore();

  useLoad(() => {
    if (isLoggedIn) {
      if (currentRole === 'receiver') {
        fetchMyParcels(true);
      } else {
        fetchTasks(undefined, true);
      }
    }
  });

  usePullDownRefresh(async () => {
    if (isLoggedIn) {
      if (currentRole === 'receiver') {
        await fetchMyParcels(true);
      } else {
        await fetchTasks(undefined, true);
      }
    }
    Taro.stopPullDownRefresh();
  });

  useReachBottom(() => {
    if (isLoggedIn && currentRole === 'receiver') {
      loadMore();
    }
  });

  // 未登录
  if (!isLoggedIn) {
    return (
      <View className='index'>
        <EmptyState
          title='请先登录'
          description='登录后查看您的包裹'
          buttonText='去登录'
          onButtonClick={() => Taro.navigateTo({ url: '/pages/login/index' })}
        />
      </View>
    );
  }

  // 跑腿员模式
  if (currentRole === 'runner') {
    return (
      <View className='index'>
        <View className='index__header'>
          <Text className='index__greeting'>你好，{userInfo ? userInfo.nickname : '跑腿员'}</Text>
          <View className='index__notify' onClick={() => Taro.navigateTo({ url: '/subpkg-user/pages/message-center/index' })}>
            <Text className='index__notify-icon'>🔔</Text>
            {unreadCount > 0 && <View className='index__badge'><Text>{unreadCount}</Text></View>}
          </View>
        </View>

        <View className='index__quick-entry'>
          <View className='index__entry-item' onClick={() => Taro.navigateTo({ url: '/subpkg-proxy/pages/proxy-hall/index' })}>
            <Text className='index__entry-icon'>📋</Text>
            <Text className='index__entry-text'>任务大厅</Text>
          </View>
          <View className='index__entry-item' onClick={() => Taro.navigateTo({ url: '/subpkg-proxy/pages/proxy-orders/index' })}>
            <Text className='index__entry-icon'>📦</Text>
            <Text className='index__entry-text'>我的订单</Text>
          </View>
        </View>

        <View className='index__section'>
          <Text className='index__section-title'>最新任务</Text>
        </View>

        {taskList.length === 0 && !loading ? (
          <EmptyState title='暂无任务' description='暂时没有待接的代取任务' />
        ) : (
          taskList.slice(0, 3).map((task) => (
            <View
              key={task.id}
              className='index__task-card'
              onClick={() => Taro.navigateTo({ url: `/subpkg-proxy/pages/proxy-detail/index?id=${task.id}` })}
            >
              <View className='index__task-header'>
                <Text className='index__task-reward'>¥{formatAmount(task.reward_amount)}</Text>
                <Text className='index__task-company'>{task.station_name}</Text>
              </View>
              <View className='index__task-body'>
                <Text className='index__task-info'>截止: {timeAgo(task.deadline)}</Text>
                {task.remark && <Text className='index__task-remark'>备注: {task.remark}</Text>}
              </View>
            </View>
          ))
        )}
      </View>
    );
  }

  // 收件人模式
  return (
    <View className='index'>
      <View className='index__header'>
        <Text className='index__greeting'>你好，{userInfo ? userInfo.nickname : '用户'}</Text>
        <View className='index__notify' onClick={() => Taro.navigateTo({ url: '/subpkg-user/pages/message-center/index' })}>
          <Text className='index__notify-icon'>🔔</Text>
          {unreadCount > 0 && <View className='index__badge'><Text>{unreadCount}</Text></View>}
        </View>
      </View>

      <View className='index__stats'>
        <View className='index__stat-item'>
          <Text className='index__stat-num'>{pendingCount}</Text>
          <Text className='index__stat-label'>待取件</Text>
        </View>
      </View>

      <View className='index__list'>
        {myParcels.length === 0 && !loading ? (
          <EmptyState title='暂无待取包裹' description='您还没有需要取的包裹' />
        ) : (
          myParcels.map((parcel) => (
            <ParcelCard
              key={parcel.id}
              parcel={parcel}
              showPickupCode
              onTap={(p) => Taro.navigateTo({ url: `/subpkg-parcel/pages/parcel-detail/index?id=${p.id}` })}
              onPickup={(p) => Taro.navigateTo({ url: `/subpkg-parcel/pages/pickup-code/index?id=${p.id}` })}
              onProxy={(p) => Taro.navigateTo({ url: `/subpkg-proxy/pages/proxy-publish/index?parcelId=${p.id}` })}
            />
          ))
        )}
      </View>
    </View>
  );
}
