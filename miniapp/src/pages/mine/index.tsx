import { View, Text, Image } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { useUserStore } from '@/stores/useUserStore';
import { maskPhone } from '@/utils/format';
import { RUNNER_STATUS } from '@/utils/constants';
import './index.scss';

export default function MinePage() {
  const { userInfo, currentRole, switchRole, logout } = useUserStore();

  const handleSwitchRole = () => {
    const newRole = currentRole === 'receiver' ? 'runner' : 'receiver';
    switchRole(newRole);
    Taro.showToast({ title: `已切换为${newRole === 'receiver' ? '收件人' : '跑腿员'}模式`, icon: 'success' });
    // 刷新首页数据
    setTimeout(() => Taro.switchTab({ url: '/pages/index/index' }), 1000);
  };

  const handleLogout = () => {
    Taro.showModal({
      title: '确认退出',
      content: '确定要退出登录吗？',
      success: (res) => {
        if (res.confirm) {
          logout();
          Taro.navigateTo({ url: '/pages/login/index' });
        }
      },
    });
  };

  const menuItems = [
    { icon: '📦', title: '我的代取订单', path: '/subpkg-proxy/pages/proxy-orders/index' },
    { icon: '🏃', title: '申请成为跑腿员', path: '/subpkg-user/pages/runner-apply/index' },
    { icon: '📍', title: '驿站导航', path: '/subpkg-user/pages/station-map/index' },
    { icon: '🔔', title: '消息中心', path: '/subpkg-user/pages/message-center/index' },
  ];

  return (
    <View className='mine-page'>
      <View className='mine-page__header'>
        <View className='mine-page__avatar'>
          <Text className='mine-page__avatar-text'>
            {userInfo ? userInfo.nickname.charAt(0) : '?'}
          </Text>
        </View>
        <View className='mine-page__info'>
          <Text className='mine-page__name'>{userInfo ? userInfo.nickname : '未登录'}</Text>
          {userInfo && (
            <>
              <Text className='mine-page__phone'>{maskPhone(userInfo.phone)}</Text>
              <View className='mine-page__credit'>
                <Text className='mine-page__credit-label'>信用分</Text>
                <Text className='mine-page__credit-value'>{userInfo.credit_score}</Text>
              </View>
            </>
          )}
        </View>
      </View>

      {/* 角色切换 */}
      {userInfo && userInfo.runner_status === RUNNER_STATUS.APPROVED && (
        <View className='mine-page__switch' onClick={handleSwitchRole}>
          <Text className='mine-page__switch-icon'>🔄</Text>
          <Text className='mine-page__switch-text'>
            切换为{currentRole === 'receiver' ? '跑腿员' : '收件人'}模式
          </Text>
          <Text className='mine-page__switch-arrow'>›</Text>
        </View>
      )}

      {/* 菜单列表 */}
      <View className='mine-page__menu'>
        {menuItems.map((item) => (
          <View
            key={item.path}
            className='mine-page__menu-item'
            onClick={() => Taro.navigateTo({ url: item.path })}
          >
            <Text className='mine-page__menu-icon'>{item.icon}</Text>
            <Text className='mine-page__menu-text'>{item.title}</Text>
            <Text className='mine-page__menu-arrow'>›</Text>
          </View>
        ))}
      </View>

      {/* 退出登录 */}
      {userInfo && (
        <View className='mine-page__logout' onClick={handleLogout}>
          <Text>退出登录</Text>
        </View>
      )}
    </View>
  );
}
