import { View, Text, Input, Image } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { useState } from 'react';
import { Button } from '@nutui/nutui-react-taro';
import { authApi } from '@/api/auth';
import { useUserStore } from '@/stores/useUserStore';
import { RUNNER_STATUS } from '@/utils/constants';
import './index.scss';

export default function RunnerApplyPage() {
  const { userInfo, updateProfile } = useUserStore();
  const [realName, setRealName] = useState('');
  const [studentId, setStudentId] = useState('');
  const [idCardImage, setIdCardImage] = useState('');
  const [loading, setLoading] = useState(false);

  const isApplied = userInfo && userInfo.runner_status !== RUNNER_STATUS.NOT_APPLIED;

  const handleChooseIdCard = async () => {
    try {
      const res = await Taro.chooseImage({
        count: 1,
        sizeType: ['compressed'],
        sourceType: ['album', 'camera'],
      });
      const filePath = res.tempFilePaths[0];
      Taro.showLoading({ title: '上传中...' });
      const result = await authApi.uploadAvatar(filePath);
      setIdCardImage(result.avatar_url);
      Taro.hideLoading();
      Taro.showToast({ title: '身份证已上传', icon: 'success' });
    } catch {
      Taro.hideLoading();
      Taro.showToast({ title: '上传失败', icon: 'none' });
    }
  };

  const handleSubmit = async () => {
    if (!realName) {
      Taro.showToast({ title: '请输入真实姓名', icon: 'none' });
      return;
    }
    if (!idCardImage) {
      Taro.showToast({ title: '请上传身份证照片', icon: 'none' });
      return;
    }

    setLoading(true);
    try {
      const res = await authApi.applyRunner({
        real_name: realName,
        student_id: studentId || undefined,
        id_card_image: idCardImage,
      });
      await updateProfile({ runner_status: res.status });
      Taro.showToast({ title: '申请已提交', icon: 'success' });
      setTimeout(() => Taro.navigateBack(), 1500);
    } catch (e: any) {
      Taro.showToast({ title: e.msg || '提交失败', icon: 'none' });
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

        <View className='runner-apply__field' onClick={handleChooseIdCard}>
          <Text className='runner-apply__label'>身份证照片</Text>
          <View className='runner-apply__idcard'>
            {idCardImage ? (
              <Image
                className='runner-apply__idcard-img'
                src={idCardImage}
                mode='aspectFill'
              />
            ) : (
              <Text className='runner-apply__idcard-placeholder'>点击上传</Text>
            )}
          </View>
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
