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
  const [deadlineLabel, setDeadlineLabel] = useState('');
  const [remark, setRemark] = useState('');
  const [loading, setLoading] = useState(false);

  const DEADLINE_OPTIONS = [
    { label: '立即', hours: 5 },
    { label: '今天', hours: 24 },
    { label: '明天', hours: 48 },
    { label: '后天', hours: 72 },
  ] as const;

  useEffect(() => {
    fetchMyParcels(true);
    const d = addHours(24);
    setDeadline(d.toISOString());
    setDeadlineLabel('今天');
  }, []);

  useEffect(() => {
    const parcelId = router.params.parcelId;
    if (parcelId && myParcels.length > 0) {
      const found = myParcels.find((p) => p.id === Number(parcelId));
      if (found) setSelectedParcel(found);
    }
  }, [router.params.parcelId, myParcels]);

  const pendingParcels = myParcels.filter((p) => p.status === 1 || p.status === 3);

  const addHours = (h: number) => {
    const d = new Date();
    d.setHours(d.getHours() + h);
    return d;
  };

  const handleSelectDeadline = (label: string, hours: number) => {
    setDeadline(addHours(hours).toISOString());
    setDeadlineLabel(label);
  };

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
      const d = new Date(deadline);
      const pad = (n: number) => String(n).padStart(2, '0');
      const deadlineStr = `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
      await publishTask({
        parcel_id: selectedParcel.id,
        reward_amount: parseFloat(reward),
        deadline: deadlineStr,
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
        <View className='proxy-publish__deadline-options'>
          {DEADLINE_OPTIONS.map((opt) => {
            const disabled = opt.label === '立即' && parseFloat(reward || '0') < 5;
            return (
              <View
                key={opt.label}
                className={`proxy-publish__deadline-btn ${deadlineLabel === opt.label ? 'proxy-publish__deadline-btn--active' : ''} ${disabled ? 'proxy-publish__deadline-btn--disabled' : ''}`}
                onClick={disabled ? undefined : () => handleSelectDeadline(opt.label, opt.hours)}
              >
                <Text>{opt.label}</Text>
              </View>
            );
          })}
        </View>
        {deadline && (() => {
          const d = new Date(deadline);
          const pad = (n: number) => String(n).padStart(2, '0');
          return (
            <Text className='proxy-publish__deadline-hint'>
              截止时间：{pad(d.getMonth() + 1)}/{pad(d.getDate())} {pad(d.getHours())}:{pad(d.getMinutes())}
            </Text>
          );
        })()}
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
