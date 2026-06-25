// app.js —— 全局应用实例
App({
  globalData: {
    // 后端基础地址，真机调试请改为可访问的域名/IP
    baseURL: 'http://127.0.0.1:8080',
    token: '',
    user: null
  },
  onLaunch() {
    const token = wx.getStorageSync('token');
    const user = wx.getStorageSync('user');
    if (token) {
      this.globalData.token = token;
      this.globalData.user = user;
    }
  }
});
