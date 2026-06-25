// pages/villager/report/report.js —— 隐患上报（文字 + 定位）
const api = require('../../../utils/request');

Page({
  data: { content: '', lng: 0, lat: 0, located: false },
  onContent(e) { this.setData({ content: e.detail.value }); },
  locate() {
    wx.getLocation({
      type: 'gcj02',
      success: (res) => this.setData({ lng: res.longitude, lat: res.latitude, located: true }),
      fail: () => wx.showToast({ title: '定位失败', icon: 'none' })
    });
  },
  submit() {
    if (!this.data.content) { wx.showToast({ title: '请填写隐患描述', icon: 'none' }); return; }
    api.post('/api/v1/reports', {
      content: this.data.content, lng: this.data.lng, lat: this.data.lat
    }).then(() => {
      wx.showToast({ title: '上报成功' });
      setTimeout(() => wx.navigateBack(), 800);
    });
  }
});
