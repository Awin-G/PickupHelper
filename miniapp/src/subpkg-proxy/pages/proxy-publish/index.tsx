import { View, Text, Input, Textarea } from '@tarojs/components';
import Taro, { useRouter } from '@tarojs/taro';
import { useState, useEffect } from 'react';
import { Button } from '@nutui/nutui-react-taro';
import { useProxyStore } from '@/stores/useProxyStore';
import { useParcelStore } from '@/stores/useParcelStore';
import { RECOMMENDED_REWARDS } from '@/utils/constants';
import { isValidAmount } from '@/utils/validator';
import type { Parcel } from '@/api/types';
import './index.scss';

export default function ProxyPublishPage() {
  const router = useRouter();
  const { publishTask } = useProxyStore();
  const { myParcels, fetchMyParcels } = useParcelStore();

  const [selectedParcel, setSelectedParcel] = useState<Parcel | null>(null);
  const [reward, setReward] = useState('5.00');
  const [deadline, setDeadline] = useState('');
  const [remark, setRemark] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    fetchMyParcels(true);
    // 默认截止时间：今天 18:00
    const now = new Date();
    const today18 = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 18, 0, 0);
    if (now > today18) {
      today18.setDate(today18.getDate() + 1);
    }
    setDeadline(today18.toISOString().slice(0, 16));
  }, []);

  useEffect(() => {
    const parcelId = router.params.parcelId;
    if (parcelId && myParcels.length > 0) {
      const found = myParcels.find((p) => p.id === Number(parcelId));
      if (found) setSelectedParcel(found);
    }
  }, [router.params.parcelId, myParcels]);

  const pendingParcels = myParcels.filter((p) => p.status === 1 || p.status === 3);

  const handleSubmit = async () => {
    if (!selectedParcel) {
      Taro.showToast({ title: '请选择包裹', icon: 'none' });
      return;
    }
    if (!isValidAmount(parseFloat(reward))) {
      Taro.showToast({ title: '悬赏金额不合法', icon: 'none' });
      return;
    }
    if (!deadline) {
      Taro.showToast({ title: '请选择截止时间', icon: 'none' });
      return;
    }

    setLoading(true);
    try {
      await publishTask({
        parcel_id: selectedParcel.id,
        reward_amount: parseFloat(reward),
        deadline: new Date(deadline).toISOString(),
        remark,
      });
      Taro.showToast({ title: '发布成功', icon: 'success' });
      setTimeout(() => {
        Taro.navigateBack();
      }, 1500);
    } catch (err: any) {
      Taro.showToast({ title: err.msg || err.message || '发布失败', icon: 'none', duration: 3000 });
    } finally {
      setLoading(false);
    }
  };

  return (
    <View className='proxy-publish'>
      <View className='proxy-publish__section'>
        <Text className='proxy-publish__label'>选择待代取包裹</Text>
        <View className='proxy-publish__parcel-list'>
          {pendingParcels.map((p) => (
            <View
              key={p.id}
              className={`proxy-publish__parcel-item ${selectedParcel && selectedParcel.id === p.id ? 'proxy-publish__parcel-item--selected' : ''}`}
              onClick={() => setSelectedParcel(p)}
            >
              <Text>{p.courier_company} | {p.shelf_code}</Text>
            </View>
          ))}
        </View>
      </View>

      <View className='proxy-publish__section'>
        <Text className='proxy-publish__label'>悬赏金额 (元)</Text>
        <Input
          className='proxy-publish__input'
          type='digit'
          value={reward}
          onInput={(e) => setReward(e.detail.value)}
        />
        <View className='proxy-publish__recommended'>
          {RECOMMENDED_REWARDS.map((r) => (
            <View
              key={r.value}
              className={`proxy-publish__recommended-item ${parseFloat(reward) === r.value ? 'proxy-publish__recommended-item--active' : ''}`}
              onClick={() => setReward(r.value.toFixed(2))}
            >
              <Text>¥{r.value}({r.label})</Text>
            </View>
          ))}
        </View>
      </View>

      <View className='proxy-publish__section'>
        <Text className='proxy-publish__label'>取件截止时间</Text>
        <Input
          className='proxy-publish__input'
          type='datetime-local'
          value={deadline}
          onInput={(e) => setDeadline(e.detail.value)}
        />
      </View>

      <View className='proxy-publish__section'>
        <Text className='proxy-publish__label'>备注 (可选)</Text>
        <Textarea
          className='proxy-publish__textarea'
          maxlength={255}
          placeholder='请输入备注...'
          value={remark}
          onInput={(e) => setRemark(e.detail.value)}
        />
      </View>

      <Button
        type='primary'
        block
        loading={loading}
        className='proxy-publish__submit'
        onClick={handleSubmit}
      >
        确认发布代取任务
      </Button>
    </View>
  );
}
