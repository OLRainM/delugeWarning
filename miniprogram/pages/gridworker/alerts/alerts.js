// pages/gridworker/alerts/alerts.js —— 预警列表（含待复核）
const api = require('../../../utils/request');

Page({
  data: { alerts: [], status: '' },
  onShow() { this.load(); },
  filter(e) { this.setData({ status: e.currentTarget.dataset.s }, () => this.load()); },
  load() {
    api.get('/api/v1/alerts', { status: this.data.status }).then((res) => {
      const items = (res.items || []).map((a) => ({
        ...a, levelInfo: api.levelMap[a.level] || { text: a.level, cls: '' }
      }));
      this.setData({ alerts: items });
    });
  },
  open(e) {
    wx.navigateTo({ url: '/pages/gridworker/alert-detail/alert-detail?id=' + e.currentTarget.dataset.id });
  }
});
