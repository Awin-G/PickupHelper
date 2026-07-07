import { View, Text, Input } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { useState } from 'react';
import { Button } from '@nutui/nutui-react-taro';
import { useUserStore } from '@/stores/useUserStore';
import { isValidPhone } from '@/utils/validator';
import { RUNNER_STATUS } from '@/utils/constants';
import './index.scss';

export default function RunnerApplyPage() {
  const { userInfo } = useUserStore();
  const [realName, setRealName] = useState('');
  const [studentId, setStudentId] = useState('');
  const [phone, setPhone] = useState('');
  const [loading, setLoading] = useState(false);

  const isApplied = userInfo && userInfo.runner_status !== RUNNER_STATUS.NOT_APPLIED;

  const handleSubmit = async () => {
    if (!realName) {
      Taro.showToast({ title: '请输入真实姓名', icon: 'none' });
      return;
    }
    if (!studentId) {
      Taro.showToast({ title: '请输入学号', icon: 'none' });
      return;
    }
    if (!isValidPhone(phone)) {
      Taro.showToast({ title: '请输入正确的手机号', icon: 'none' });
      return;
    }

    setLoading(true);
    try {
      // Mock 提交
      await new Promise((r) => setTimeout(r, 1000));
      Taro.showToast({ title: '申请已提交', icon: 'success' });
      setTimeout(() => Taro.navigateBack(), 1500);
    } catch {
      Taro.showToast({ title: '提交失败', icon: 'none' });
    } finally {
      setLoading(false);
    }
  };

  if (isApplied) {
    const statusText: Record<number, string> = {
      [RUNNER_STATUS.PENDING]: '审核中，请耐心等待',
      [RUNNER_STATUS.APPROVED]: '您已是跑腿员',
      [RUNNER_STATUS.REJECTED]: '申请被拒绝，请重新申请',
    };
    return (
      <View className='runner-apply'>
        <View className='runner-apply__status'>
          <Text className='runner-apply__status-text'>{statusText[userInfo.runner_status]}</Text>
        </View>
      </View>
    );
  }

  return (
    <View className='runner-apply'>
      <View className='runner-apply__form'>
        <View className='runner-apply__field'>
          <Text className='runner-apply__label'>真实姓名</Text>
          <Input
            className='runner-apply__input'
            placeholder='请输入真实姓名'
            value={realName}
            onInput={(e) => setRealName(e.detail.value)}
          />
        </View>

        <View className='runner-apply__field'>
          <Text className='runner-apply__label'>学号</Text>
          <Input
            className='runner-apply__input'
            placeholder='请输入学号'
            value={studentId}
            onInput={(e) => setStudentId(e.detail.value)}
          />
        </View>

        <View className='runner-apply__field'>
          <Text className='runner-apply__label'>手机号</Text>
          <Input
            className='runner-apply__input'
            type='number'
            maxlength={11}
            placeholder='请输入手机号'
            value={phone}
            onInput={(e) => setPhone(e.detail.value)}
          />
        </View>
      </View>

      <Button
        type='primary'
        block
        loading={loading}
        className='runner-apply__submit'
        onClick={handleSubmit}
      >
        提交申请
      </Button>
    </View>
  );
}
