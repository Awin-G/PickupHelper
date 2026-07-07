import { View, Text, Input, Image } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { useState } from 'react';
import { useUserStore } from '@/stores/useUserStore';
import { authApi } from '@/api/auth';
import { storage } from '@/utils/storage';
import './index.scss';

export default function ProfileEditPage() {
  const { userInfo, updateProfile } = useUserStore();
  const [nickname, setNickname] = useState(userInfo?.nickname || '');
  const [saving, setSaving] = useState(false);

  const handleChooseAvatar = async () => {
    try {
      const res = await Taro.chooseImage({
        count: 1,
        sizeType: ['compressed'],
        sourceType: ['album', 'camera'],
      });
      const filePath = res.tempFilePaths[0];

      Taro.showLoading({ title: '上传中...' });
      const result = await authApi.uploadAvatar(filePath);
      await updateProfile({ avatar: result.avatar_url });
      Taro.hideLoading();
      Taro.showToast({ title: '头像已更新', icon: 'success' });
    } catch {
      Taro.hideLoading();
      Taro.showToast({ title: '上传失败', icon: 'none' });
    }
  };

  const handleSaveNickname = async () => {
    if (!nickname.trim()) {
      Taro.showToast({ title: '昵称不能为空', icon: 'none' });
      return;
    }
    if (nickname.length > 50) {
      Taro.showToast({ title: '昵称最长50字符', icon: 'none' });
      return;
    }

    setSaving(true);
    try {
      await updateProfile({ nickname: nickname.trim() });
      Taro.showToast({ title: '保存成功', icon: 'success' });
    } catch {
      Taro.showToast({ title: '保存失败', icon: 'none' });
    } finally {
      setSaving(false);
    }
  };

  const getAvatarUrl = () => {
    if (userInfo?.avatar) {
      // 如果是相对路径，加上 API 前缀
      if (userInfo.avatar.startsWith('/')) {
        const base = process.env.TARO_APP_API_BASE || 'http://localhost:18080/api/v1';
        return base.replace('/api/v1', '') + userInfo.avatar;
      }
      return userInfo.avatar;
    }
    return '';
  };

  return (
    <View className='profile-edit'>
      <View className='profile-edit__avatar-section' onClick={handleChooseAvatar}>
        <View className='profile-edit__avatar'>
          {userInfo?.avatar ? (
            <Image className='profile-edit__avatar-img' src={getAvatarUrl()} mode='aspectFill' />
          ) : (
            <Text className='profile-edit__avatar-text'>
              {userInfo?.nickname ? userInfo.nickname.charAt(0) : '?'}
            </Text>
          )}
        </View>
        <Text className='profile-edit__avatar-hint'>点击更换头像</Text>
      </View>

      <View className='profile-edit__form'>
        <View className='profile-edit__field'>
          <Text className='profile-edit__label'>昵称</Text>
          <Input
            className='profile-edit__input'
            maxlength={50}
            placeholder='请输入昵称'
            value={nickname}
            onInput={(e) => setNickname(e.detail.value)}
          />
        </View>

        <View className='profile-edit__field'>
          <Text className='profile-edit__label'>手机号</Text>
          <Text className='profile-edit__value'>{userInfo?.phone || ''}</Text>
        </View>

        <View className='profile-edit__field'>
          <Text className='profile-edit__label'>信用分</Text>
          <Text className='profile-edit__value'>{userInfo?.credit_score || 0}</Text>
        </View>
      </View>

      <View
        className={`profile-edit__save ${saving ? 'profile-edit__save--loading' : ''}`}
        onClick={saving ? undefined : handleSaveNickname}
      >
        <Text>{saving ? '保存中...' : '保存昵称'}</Text>
      </View>
    </View>
  );
}
