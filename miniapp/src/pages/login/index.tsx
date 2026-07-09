import { View, Text, Input, Button as TaroButton } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { useState, useCallback, useRef } from 'react';
import { Button } from '@nutui/nutui-react-taro';
import { useUserStore } from '@/stores/useUserStore';
import { authApi } from '@/api/auth';
import { generateNickname } from '@/utils/format';
import { isValidPhone, isValidCode } from '@/utils/validator';
import { ENABLE_WECHAT_LOGIN } from '@/config';
import { storage } from '@/utils/storage';
import './index.scss';

export default function LoginPage() {
  const [phone, setPhone] = useState('');
  const [code, setCode] = useState('');
  const [countdown, setCountdown] = useState(0);
  const [loading, setLoading] = useState(false);
  const [wxLoading, setWxLoading] = useState(false);
  const [agreed, setAgreed] = useState(false);
  const [wxCode, setWxCode] = useState('');
  const timerRef = useRef<NodeJS.Timeout | null>(null);

  const { login, isLoggedIn } = useUserStore();

  if (isLoggedIn) {
    Taro.switchTab({ url: '/pages/index/index' });
    return null;
  }

  // 提前获取 wx.login code（仅微信登录启用时）
  if (ENABLE_WECHAT_LOGIN && !wxCode) {
    Taro.login().then((res) => {
      if (res.code) setWxCode(res.code);
    });
  }

  const startCountdown = useCallback(() => {
    setCountdown(60);
    timerRef.current = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          if (timerRef.current) clearInterval(timerRef.current);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
  }, []);

  const handleSendCode = async () => {
    if (!isValidPhone(phone)) {
      Taro.showToast({ title: '请输入正确的手机号', icon: 'none' });
      return;
    }
    try {
      await authApi.sendCode(phone);
      startCountdown();
      Taro.showToast({ title: '验证码已发送', icon: 'success' });
    } catch {
      Taro.showToast({ title: '发送失败，请重试', icon: 'none' });
    }
  };

  const handleLogin = async () => {
    if (!isValidPhone(phone)) {
      Taro.showToast({ title: '请输入正确的手机号', icon: 'none' });
      return;
    }
    if (!isValidCode(code)) {
      Taro.showToast({ title: '请输入验证码', icon: 'none' });
      return;
    }
    if (!agreed) {
      Taro.showToast({ title: '请同意用户协议', icon: 'none' });
      return;
    }

    setLoading(true);
    try {
      await login(phone, code);
      Taro.showToast({ title: '登录成功', icon: 'success' });
      setTimeout(() => Taro.switchTab({ url: '/pages/index/index' }), 1000);
    } catch {
      Taro.showToast({ title: '登录失败，请重试', icon: 'none' });
    } finally {
      setLoading(false);
    }
  };

  // 微信手机号授权回调
  const handleGetPhoneNumber = async (e: any) => {
    if (!agreed) {
      Taro.showToast({ title: '请同意用户协议', icon: 'none' });
      return;
    }

    if (e.detail.errMsg !== 'getPhoneNumber:ok') {
      Taro.showToast({ title: '手机号授权取消', icon: 'none' });
      return;
    }

    const phoneCode = e.detail.code;
    if (!phoneCode) {
      Taro.showToast({ title: '获取手机号失败', icon: 'none' });
      return;
    }

    if (!wxCode) {
      Taro.showToast({ title: '微信登录凭证获取中，请稍后重试', icon: 'none' });
      Taro.login().then((res) => {
        if (res.code) setWxCode(res.code);
      });
      return;
    }

    setWxLoading(true);
    try {
      const nickname = generateNickname();
      const result = await authApi.wechatLogin({
        code: wxCode,
        phone_code: phoneCode,
        nickname,
      });

      storage.set('token', result.access_token);
      storage.set('refresh_token', result.refresh_token);
      useUserStore.setState({
        token: result.access_token,
        refreshToken: result.refresh_token,
        userInfo: result.user,
        isLoggedIn: true,
      });

      Taro.showToast({ title: '登录成功', icon: 'success' });
      setTimeout(() => Taro.switchTab({ url: '/pages/index/index' }), 1000);
    } catch (err: any) {
      Taro.showToast({ title: `登录失败: ${err.msg || '未知错误'}`, icon: 'none', duration: 3000 });
      Taro.login().then((res) => {
        if (res.code) setWxCode(res.code);
      });
    } finally {
      setWxLoading(false);
    }
  };

  return (
    <View className='login-page'>
      <View className='login-page__header'>
        <Text className='login-page__title'>快递驿站助手</Text>
        <Text className='login-page__subtitle'>便捷取件，轻松代取</Text>
      </View>

      <View className='login-page__form'>
        {/* 微信手机号授权登录（仅企业主体小程序可用） */}
        {ENABLE_WECHAT_LOGIN && (
          <TaroButton
            className='login-page__wx-phone-btn'
            open-type='getPhoneNumber'
            onGetPhoneNumber={handleGetPhoneNumber}
          >
            <Text className='login-page__wx-icon'>📱</Text>
            <Text className='login-page__wx-text'>{wxLoading ? '登录中...' : '微信手机号快捷登录'}</Text>
          </TaroButton>
        )}

        {ENABLE_WECHAT_LOGIN && (
          <View className='login-page__divider'>
            <Text className='login-page__divider-text'>或</Text>
          </View>
        )}

        {/* 手机号验证码登录 */}
        <View className='login-page__input-group'>
          <Text className='login-page__prefix'>+86</Text>
          <Input
            className='login-page__input'
            type='number'
            maxlength={11}
            placeholder='请输入手机号'
            value={phone}
            onInput={(e) => setPhone(e.detail.value)}
          />
        </View>

        <View className='login-page__input-group'>
          <Input
            className='login-page__input login-page__input--code'
            type='number'
            maxlength={6}
            placeholder='请输入验证码'
            value={code}
            onInput={(e) => setCode(e.detail.value)}
          />
          <View
            className={`login-page__code-btn ${countdown > 0 ? 'login-page__code-btn--disabled' : ''}`}
            onClick={countdown > 0 ? undefined : handleSendCode}
          >
            <Text>{countdown > 0 ? `${countdown}s` : '获取验证码'}</Text>
          </View>
        </View>

        <Button
          type='primary'
          block
          loading={loading}
          className='login-page__submit'
          onClick={handleLogin}
        >
          登 录
        </Button>

        <View className='login-page__agreement' onClick={() => setAgreed(!agreed)}>
          <View className={`login-page__checkbox ${agreed ? 'login-page__checkbox--checked' : ''}`}>
            {agreed && <Text className='login-page__checkmark'>✓</Text>}
          </View>
          <Text className='login-page__agreement-text'>
            已阅读并同意《用户协议》和《隐私政策》
          </Text>
        </View>
      </View>
    </View>
  );
}
