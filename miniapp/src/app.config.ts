export default defineAppConfig({
  pages: [
    'pages/index/index',
    'pages/mine/index',
    'pages/login/index',
  ],
  subPackages: [
    {
      root: 'subpkg-parcel',
      pages: [
        'pages/parcel-detail/index',
        'pages/pickup-code/index',
        'pages/self-checkout/index',
      ],
    },
    {
      root: 'subpkg-proxy',
      pages: [
        'pages/proxy-publish/index',
        'pages/proxy-hall/index',
        'pages/proxy-detail/index',
        'pages/proxy-orders/index',
      ],
    },
    {
      root: 'subpkg-user',
      pages: [
        'pages/runner-apply/index',
        'pages/message-center/index',
        'pages/station-map/index',
      ],
    },
  ],
  preloadRule: {
    'pages/index/index': {
      network: 'all',
      packages: ['subpkg-parcel'],
    },
    'pages/mine/index': {
      network: 'all',
      packages: ['subpkg-user', 'subpkg-proxy'],
    },
  },
  tabBar: {
    color: '#999999',
    selectedColor: '#1890FF',
    list: [
      {
        pagePath: 'pages/index/index',
        text: '包裹',
        iconPath: 'assets/icons/parcel.png',
        selectedIconPath: 'assets/icons/parcel-active.png',
      },
      {
        pagePath: 'pages/mine/index',
        text: '我的',
        iconPath: 'assets/icons/mine.png',
        selectedIconPath: 'assets/icons/mine-active.png',
      },
    ],
  },
  window: {
    navigationBarTitleText: '快递驿站助手',
    navigationBarBackgroundColor: '#1890FF',
    navigationBarTextStyle: 'white',
    backgroundTextStyle: 'light',
  },
});
