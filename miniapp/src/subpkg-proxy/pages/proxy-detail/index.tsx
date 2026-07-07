import { View, Text } from '@tarojs/components';
import Taro, { useRouter } from '@tarojs/taro';
import { useState, useEffect } from 'react';
import { Button } from '@nutui/nutui-react-taro';
import { proxyApi } from '@/api/proxy';
import StatusBadge from '@/components/StatusBadge';
import { formatAmount, formatTime } from '@/utils/format';
import { PROXY_STATUS } from '@/utils/constants';
import type { ProxyOrder } from '@/api/types';
import './index.scss';

export default function ProxyDetailPage() {
  const router = useRouter();
  const [order, setOrder] = useState<ProxyOrder | null>(null);

  useEffect(() => {
    const id = router.params.id;
    if (id) {
      proxyApi.getDetail(Number(id)).then(setOrder);
    }
  }, [router.params.id]);

  if (!order) {
    return <View className='proxy-detail'><Text>加载中...</Text></View>;
  }

  const isDelivering = order.status === PROXY_STATUS.DELIVERING;
  const isConfirming = order.status === PROXY_STATUS.CONFIRMING;

  const handleConfirmDelivery = async () => {
    Taro.showModal({
      title: '确认送达',
      content: '确认已将包裹送到收件人手中？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await proxyApi.confirmDelivery(order.id, { accepted: true });
            Taro.showToast({ title: '已确认送达', icon: 'success' });
            setOrder({ ...order, status: PROXY_STATUS.COMPLETED, status_text: '已完成' });
          } catch {
            Taro.showToast({ title: '操作失败', icon: 'none' });
          }
        }
      },
    });
  };

  const handleAccept = async () => {
    Taro.showModal({
      title: '确认收货',
      content: '确认收到包裹？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await proxyApi.confirmDelivery(order.id, { accepted: true });
            Taro.showToast({ title: '已确认收货', icon: 'success' });
            setOrder({ ...order, status: PROXY_STATUS.COMPLETED, status_text: '已完成' });
          } catch {
            Taro.showToast({ title: '操作失败', icon: 'none' });
          }
        }
      },
    });
  };

  const handleReject = async () => {
    Taro.showModal({
      title: '拒绝收货',
      content: '请输入拒绝原因',
      editable: true,
      success: async (res) => {
        if (res.confirm) {
          try {
            await proxyApi.confirmDelivery(order.id, { accepted: false, reason: res.content });
            Taro.showToast({ title: '已拒绝', icon: 'success' });
          } catch {
            Taro.showToast({ title: '操作失败', icon: 'none' });
          }
        }
      },
    });
  };

  return (
    <View className='proxy-detail'>
      <View className='proxy-detail__header'>
        <StatusBadge type='proxy' status={order.status} />
      </View>

      <View className='proxy-detail__card'>
        <View className='proxy-detail__row'>
          <Text className='proxy-detail__label'>驿站</Text>
          <Text className='proxy-detail__value'>{order.station_name}</Text>
        </View>
        <View className='proxy-detail__row'>
          <Text className='proxy-detail__label'>悬赏金额</Text>
          <Text className='proxy-detail__value proxy-detail__value--reward'>¥{formatAmount(order.reward_amount)}</Text>
        </View>
        <View className='proxy-detail__row'>
          <Text className='proxy-detail__label'>截止时间</Text>
          <Text className='proxy-detail__value'>{formatTime(order.deadline)}</Text>
        </View>
        {order.temp_pickup_code && (
          <View className='proxy-detail__row'>
            <Text className='proxy-detail__label'>临时取件码</Text>
            <Text className='proxy-detail__value proxy-detail__value--code'>{order.temp_pickup_code}</Text>
          </View>
        )}
        {order.delivery_time && (
          <View className='proxy-detail__row'>
            <Text className='proxy-detail__label'>送达时间</Text>
            <Text className='proxy-detail__value'>{formatTime(order.delivery_time)}</Text>
          </View>
        )}
      </View>

      {/* 跑腿员视角 - 配送中 */}
      {isDelivering && (
        <View className='proxy-detail__actions'>
          <Button type='primary' block onClick={handleConfirmDelivery}>确认送达</Button>
        </View>
      )}

      {/* 收件人视角 - 待确认 */}
      {isConfirming && (
        <View className='proxy-detail__actions'>
          <Button type='primary' block onClick={handleAccept}>确认收货</Button>
          <Button block className='proxy-detail__btn--outline' onClick={handleReject}>拒绝</Button>
        </View>
      )}
    </View>
  );
}
