// pages/gridworker/mine/mine.js
Page({
  data: { user: {} },
  onShow() { this.setData({ user: getApp().globalData.user || {} }); },
  logout() {
    wx.removeStorageSync('token');
    wx.removeStorageSync('user');
    getApp().globalData.token = '';
    getApp().globalData.user = null;
    wx.reLaunch({ url: '/pages/login/login' });
  }
});
