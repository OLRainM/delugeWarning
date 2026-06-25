// pages/villager/home/home.js —— 村民首页：本村生效预警
const api = require('../../../utils/request');

Page({
  data: { alerts: [], loading: true },
  onShow() { this.load(); },
  load() {
    api.get('/api/v1/village/alerts').then((res) => {
      const items = (res.items || []).map((a) => ({
        ...a, levelInfo: api.levelMap[a.level] || { text: a.level, cls: '' }
      }));
      this.setData({ alerts: items, loading: false });
    }).catch(() => this.setData({ loading: false }));
  },
  openBroadcast(e) {
    wx.navigateTo({ url: '/pages/villager/broadcast/broadcast?id=' + e.currentTarget.dataset.id });
  }
});
