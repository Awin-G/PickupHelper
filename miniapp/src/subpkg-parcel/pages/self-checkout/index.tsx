import { View, Text, Input } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { useState } from 'react';
import { Button } from '@nutui/nutui-react-taro';
import './index.scss';

export default function SelfCheckoutPage() {
  const [pickupCode, setPickupCode] = useState('');
  const [loading, setLoading] = useState(false);

  const handleScan = () => {
    Taro.scanCode({
      success: (res) => {
        setPickupCode(res.result);
      },
      fail: () => {
        Taro.showToast({ title: '扫码取消', icon: 'none' });
      },
    });
  };

  const handleSubmit = async () => {
    if (!pickupCode || pickupCode.length < 4) {
      Taro.showToast({ title: '请输入或扫描取件码', icon: 'none' });
      return;
    }

    setLoading(true);
    try {
      // Mock 提交
      await new Promise((r) => setTimeout(r, 1000));
      Taro.showToast({ title: '出库成功', icon: 'success' });
      setTimeout(() => Taro.navigateBack(), 1500);
    } catch {
      Taro.showToast({ title: '出库失败', icon: 'none' });
    } finally {
      setLoading(false);
    }
  };

  return (
    <View className='self-checkout'>
      <View className='self-checkout__card'>
        <Text className='self-checkout__title'>自助出库</Text>
        <Text className='self-checkout__desc'>扫描驿站二维码或输入取件码完成出库</Text>

        <View className='self-checkout__input-group'>
          <Input
            className='self-checkout__input'
            type='number'
            placeholder='请输入取件码'
            value={pickupCode}
            onInput={(e) => setPickupCode(e.detail.value)}
          />
        </View>

        <View className='self-checkout__divider'>
          <Text className='self-checkout__divider-text">或</Text>
        </View>

        <Button
          block
          className='self-checkout__scan-btn'
          onClick={handleScan}
        >
          扫描驿站二维码
        </Button>
      </View>

      <Button
        type='primary'
        block
        loading={loading}
        className='self-checkout__submit'
        onClick={handleSubmit}
      >
        确认出库
      </Button>
    </View>
  );
}
