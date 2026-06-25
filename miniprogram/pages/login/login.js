// pages/login/login.js
const api = require('../../utils/request');

Page({
  data: { role: 'villager' },
  chooseRole(e) {
    this.setData({ role: e.currentTarget.dataset.role });
  },
  login() {
    const role = this.data.role;
    wx.login({
      success: (res) => {
        api.post('/api/v1/auth/wx-login', { code: res.code, role }).then((data) => {
          const app = getApp();
          app.globalData.token = data.token;
          app.globalData.user = data.user;
          wx.setStorageSync('token', data.token);
          wx.setStorageSync('user', data.user);
          if (data.user.role === 'gridworker') {
            wx.reLaunch({ url: '/pages/gridworker/tasks/tasks' });
          } else {
            wx.reLaunch({ url: '/pages/villager/home/home' });
          }
        });
      }
    });
  }
});
