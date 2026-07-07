import { View, Text } from '@tarojs/components';
import { useLoad, usePullDownRefresh, useReachBottom, useRouter } from '@tarojs/taro';
import { useUserStore } from '@/stores/useUserStore';
import { useParcelStore } from '@/stores/useParcelStore';
import ParcelCard from '@/components/ParcelCard';
import EmptyState from '@/components/EmptyState';
import './index.scss';

export default function Index() {
  const { isLoggedIn, currentRole } = useUserStore();
  const { myParcels, pendingCount, loading, hasMore, fetchMyParcels, loadMore } = useParcelStore();

  useLoad(() => {
    if (isLoggedIn) {
      fetchMyParcels(true);
    }
  });

  usePullDownRefresh(async () => {
    if (isLoggedIn) {
      await fetchMyParcels(true);
    }
  });

  useReachBottom(() => {
    if (isLoggedIn) {
      loadMore();
    }
  });

  if (!isLoggedIn) {
    return (
      <View className='index'>
        <EmptyState
          title='请先登录'
          description='登录后查看您的包裹'
          buttonText='去登录'
          onButtonClick={() => {
            Taro.navigateTo({ url: '/pages/login/index' });
          }}
        />
      </View>
    );
  }

  return (
    <View className='index'>
      {/* 统计卡片 */}
      <View className='index__stats'>
        <View className='index__stat-item'>
          <Text className='index__stat-num'>{pendingCount}</Text>
          <Text className='index__stat-label'>待取件</Text>
        </View>
      </View>

      {/* 包裹列表 */}
      <View className='index__list'>
        {myParcels.length === 0 && !loading ? (
          <EmptyState
            title='暂无待取包裹'
            description='您还没有需要取的包裹'
          />
        ) : (
          myParcels.map((parcel) => (
            <ParcelCard
              key={parcel.id}
              parcel={parcel}
              showPickupCode
              onTap={(p) => {
                Taro.navigateTo({ url: `/subpkg-parcel/pages/parcel-detail/index?id=${p.id}` });
              }}
              onPickup={(p) => {
                Taro.navigateTo({ url: `/subpkg-parcel/pages/pickup-code/index?id=${p.id}` });
              }}
              onProxy={(p) => {
                Taro.navigateTo({ url: `/subpkg-proxy/pages/proxy-publish/index?parcelId=${p.id}` });
              }}
            />
          ))
        )}
      </View>
    </View>
  );
}
