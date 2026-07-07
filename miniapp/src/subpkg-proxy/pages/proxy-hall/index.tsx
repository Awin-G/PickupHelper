import { View, Text } from '@tarojs/components';
import Taro, { usePullDownRefresh, useReachBottom } from '@tarojs/taro';
import { useState, useEffect } from 'react';
import { Button } from '@nutui/nutui-react-taro';
import { useProxyStore } from '@/stores/useProxyStore';
import EmptyState from '@/components/EmptyState';
import { formatAmount, timeAgo } from '@/utils/format';
import './index.scss';

export default function ProxyHallPage() {
  const { taskList, taskLoading, taskHasMore, fetchTasks, loadMoreTasks, acceptTask } = useProxyStore();
  const [sortBy, setSortBy] = useState<'reward' | 'created_at' | 'deadline'>('created_at');

  useEffect(() => {
    fetchTasks({ sort_by: sortBy }, true);
  }, [sortBy]);

  usePullDownRefresh(async () => {
    await fetchTasks({ sort_by: sortBy }, true);
    Taro.stopPullDownRefresh();
  });

  useReachBottom(() => {
    loadMoreTasks();
  });

  const handleAccept = (id: number) => {
    Taro.showModal({
      title: '确认接单',
      content: '确认接下此代取任务？',
      success: async (res) => {
        if (res.confirm) {
          try {
            const order = await acceptTask(id);
            Taro.showToast({ title: '接单成功', icon: 'success' });
            Taro.navigateTo({ url: `/subpkg-proxy/pages/proxy-detail/index?id=${order.id}` });
          } catch {
            Taro.showToast({ title: '接单失败', icon: 'none' });
          }
        }
      },
    });
  };

  return (
    <View className='proxy-hall'>
      <View className='proxy-hall__sort'>
        <View
          className={`proxy-hall__sort-btn ${sortBy === 'reward' ? 'proxy-hall__sort-btn--active' : ''}`}
          onClick={() => setSortBy('reward')}
        >
          <Text>悬赏最高</Text>
        </View>
        <View
          className={`proxy-hall__sort-btn ${sortBy === 'created_at' ? 'proxy-hall__sort-btn--active' : ''}`}
          onClick={() => setSortBy('created_at')}
        >
          <Text>最新发布</Text>
        </View>
        <View
          className={`proxy-hall__sort-btn ${sortBy === 'deadline' ? 'proxy-hall__sort-btn--active' : ''}`}
          onClick={() => setSortBy('deadline')}
        >
          <Text>截止最近</Text>
        </View>
      </View>

      <View className='proxy-hall__list'>
        {taskList.length === 0 && !taskLoading ? (
          <EmptyState title='暂无任务' description='暂时没有待接的代取任务' />
        ) : (
          taskList.map((task) => (
            <View key={task.id} className='proxy-hall__card'>
              <View className='proxy-hall__card-header'>
                <Text className='proxy-hall__reward'>¥{formatAmount(task.reward_amount)}</Text>
                <Text className='proxy-hall__station'>{task.station_name}</Text>
              </View>
              <View className='proxy-hall__card-body'>
                <Text className='proxy-hall__deadline'>截止: {timeAgo(task.deadline)}</Text>
                {task.remark && (
                  <Text className='proxy-hall__remark'>备注: {task.remark}</Text>
                )}
              </View>
              <View className='proxy-hall__card-footer'>
                <Button
                  type='primary'
                  size='small'
                  className='proxy-hall__accept-btn'
                  onClick={() => handleAccept(task.id)}
                >
                  接单
                </Button>
              </View>
            </View>
          ))
        )}
      </View>
    </View>
  );
}
