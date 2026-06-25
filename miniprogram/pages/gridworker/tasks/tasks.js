// pages/gridworker/tasks/tasks.js —— 网格员任务列表
const api = require('../../../utils/request');

Page({
  data: { tasks: [], status: '' },
  onShow() { this.load(); },
  filter(e) { this.setData({ status: e.currentTarget.dataset.s }, () => this.load()); },
  load() {
    api.get('/api/v1/tasks', { status: this.data.status }).then((res) => {
      this.setData({ tasks: res.items || [] });
    });
  },
  open(e) {
    wx.navigateTo({ url: '/pages/gridworker/task-detail/task-detail?id=' + e.currentTarget.dataset.id });
  },
  goAlerts() { wx.navigateTo({ url: '/pages/gridworker/alerts/alerts' }); },
  goMine() { wx.navigateTo({ url: '/pages/gridworker/mine/mine' }); }
});
